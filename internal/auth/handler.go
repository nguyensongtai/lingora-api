package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/nguyensongtai/lingora-api/internal/api"
	"github.com/nguyensongtai/lingora-api/internal/httpx"
	"github.com/nguyensongtai/lingora-api/internal/ratelimit"
	"github.com/nguyensongtai/lingora-api/internal/user"
)

// service là những gì Handler cần ở tầng nghiệp vụ.
type service interface {
	Register(ctx context.Context, email, password, displayName string) (TokenPair, error)
	SignInWithGoogle(ctx context.Context, code, redirectURI string) (TokenPair, error)
	Login(ctx context.Context, email, password string) (TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (TokenPair, error)
	Logout(ctx context.Context, refreshToken string) error
	CurrentUser(ctx context.Context, userID string) (user.User, error)
}

// LoginIPRule giới hạn số lần gọi các route xác thực từ một địa chỉ IP.
// Rộng hơn hạn mức theo email vì một văn phòng hay quán cà phê dùng chung IP.
var LoginIPRule = ratelimit.Rule{
	Name:   "auth-ip",
	Limit:  20,
	Window: 15 * time.Minute,
}

// Handler ánh xạ HTTP sang nghiệp vụ xác thực.
type Handler struct {
	service service
	limiter ratelimit.Limiter
}

// NewHandler nhận service đã sẵn sàng dùng. limiter nil thì dùng bộ đếm trong
// bộ nhớ — không có đường nào tắt hẳn hạn mức.
func NewHandler(svc service, limiter ratelimit.Limiter) *Handler {
	if limiter == nil {
		limiter = ratelimit.NewMemory()
	}
	return &Handler{service: svc, limiter: limiter}
}

// Mount gắn route xác thực; authenticated bọc route cần token hợp lệ.
func (h *Handler) Mount(r chi.Router, authenticated func(http.Handler) http.Handler) {
	r.Route("/auth", func(r chi.Router) {
		// Chỉ bọc những route mở phiên mới. /refresh và /logout đã cầm sẵn một
		// token hợp lệ nên không phải cửa để dò, còn chặn chúng thì một người
		// dùng bình thường mở nhiều tab sẽ bị khoá oan.
		r.Group(func(r chi.Router) {
			r.Use(httpx.RateLimitByIP(h.limiter, LoginIPRule))
			r.Post("/register", h.register)
			r.Post("/login", h.login)
			r.Post("/google", h.google)
		})

		r.Post("/refresh", h.refresh)
		r.Post("/logout", h.logout)
		r.With(authenticated).Get("/me", h.me)
	})
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var body api.RegisterRequest
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		writeMalformed(w, err)
		return
	}

	pair, err := h.service.Register(r.Context(), body.Email, body.Password, body.DisplayName)
	if err != nil {
		writeError(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, toAPITokenPair(pair))
}

func (h *Handler) google(w http.ResponseWriter, r *http.Request) {
	var body api.GoogleSignInRequest
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		writeMalformed(w, err)
		return
	}

	pair, err := h.service.SignInWithGoogle(r.Context(), body.Code, body.RedirectUri)
	if err != nil {
		writeError(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, toAPITokenPair(pair))
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var body api.LoginRequest
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		writeMalformed(w, err)
		return
	}

	missing := map[string]string{}
	if strings.TrimSpace(body.Email) == "" {
		missing["email"] = "không được rỗng"
	}
	if body.Password == "" {
		missing["password"] = "không được rỗng"
	}
	if len(missing) > 0 {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidation, "Dữ liệu không hợp lệ.", missing)
		return
	}

	pair, err := h.service.Login(r.Context(), body.Email, body.Password)
	if err != nil {
		writeError(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, toAPITokenPair(pair))
}

func (h *Handler) refresh(w http.ResponseWriter, r *http.Request) {
	body, ok := decodeRefreshRequest(w, r)
	if !ok {
		return
	}

	pair, err := h.service.Refresh(r.Context(), body.RefreshToken)
	if err != nil {
		writeError(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, toAPITokenPair(pair))
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	body, ok := decodeRefreshRequest(w, r)
	if !ok {
		return
	}

	if err := h.service.Logout(r.Context(), body.RefreshToken); err != nil {
		writeError(w, r, err)
		return
	}
	httpx.NoContent(w)
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	claims, ok := ClaimsFrom(r.Context())
	if !ok {
		// Middleware luôn gắn claims, nên tới đây là lỗi lập trình chứ không
		// phải lỗi của client.
		slog.ErrorContext(r.Context(), "claims missing behind the auth middleware")
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternal, "Có lỗi xảy ra, vui lòng thử lại.", nil)
		return
	}

	account, err := h.service.CurrentUser(r.Context(), claims.Subject)
	if err != nil {
		writeError(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, toAPIUser(account))
}

func decodeRefreshRequest(w http.ResponseWriter, r *http.Request) (api.RefreshRequest, bool) {
	var body api.RefreshRequest
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		writeMalformed(w, err)
		return api.RefreshRequest{}, false
	}
	if strings.TrimSpace(body.RefreshToken) == "" {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidation, "Dữ liệu không hợp lệ.", map[string]string{
			"refresh_token": "không được rỗng",
		})
		return api.RefreshRequest{}, false
	}
	return body, true
}

func toAPIUser(account user.User) api.User {
	return api.User{
		Id:          account.ID,
		Email:       account.Email,
		DisplayName: account.DisplayName,
		Role:        api.UserRole(account.Role),
		CreatedAt:   account.CreatedAt,
	}
}

func toAPITokenPair(pair TokenPair) api.TokenPair {
	return api.TokenPair{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    pair.ExpiresIn,
		User:         toAPIUser(pair.User),
	}
}

func writeMalformed(w http.ResponseWriter, err error) {
	httpx.Error(w, http.StatusBadRequest, httpx.CodeMalformed, "Body không đọc được.", map[string]string{
		"body": err.Error(),
	})
}

// writeError dịch sentinel sang mã HTTP. Sai thông tin đăng nhập và phiên hết
// hạn đều là 401 để client chỉ cần xử lý một trường hợp: đăng nhập lại.
func writeError(w http.ResponseWriter, r *http.Request, err error) {
	var validationErr *ValidationError

	var rateLimitedErr *RateLimitedError

	switch {
	case errors.As(err, &rateLimitedErr):
		httpx.WriteRateLimited(w, rateLimitedErr.RetryAfter)
	case errors.As(err, &validationErr):
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidation, "Dữ liệu không hợp lệ.", validationErr.Fields)
	case errors.Is(err, ErrGoogleNotConfigured):
		httpx.Error(w, http.StatusServiceUnavailable, httpx.CodeInternal, "Đăng nhập bằng Google chưa được bật.", nil)
	case errors.Is(err, ErrGoogleEmailUnverified):
		httpx.Error(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "Google chưa xác minh email của tài khoản này.", nil)
	case errors.Is(err, ErrGoogleExchangeFailed):
		// Lý do cụ thể đã nằm trong log; với client thì mọi trường hợp đều là
		// "phiên đăng nhập Google không dùng được nữa".
		slog.WarnContext(r.Context(), "google sign-in rejected", slog.Any("error", err))
		httpx.Error(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "Không xác thực được với Google, vui lòng thử lại.", nil)
	case errors.Is(err, user.ErrGoogleAccountTaken), errors.Is(err, user.ErrGoogleAlreadyLinked):
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "Tài khoản Google này đã gắn với một tài khoản khác.", nil)
	case errors.Is(err, user.ErrEmailTaken):
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "Email này đã có tài khoản.", map[string]string{
			"email": "đã được đăng ký",
		})
	case errors.Is(err, ErrInvalidCredentials):
		httpx.Error(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "Email hoặc mật khẩu không đúng.", nil)
	case errors.Is(err, ErrSessionNotFound):
		httpx.Error(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "Phiên đăng nhập đã hết hạn, vui lòng đăng nhập lại.", nil)
	case errors.Is(err, user.ErrNotFound):
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "Không tìm thấy tài khoản.", nil)
	case errors.Is(err, user.ErrInvalidID), errors.Is(err, ErrInvalidUserID):
		httpx.Error(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "Access token không hợp lệ.", nil)
	default:
		slog.ErrorContext(r.Context(), "auth request failed", slog.Any("error", err))
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternal, "Có lỗi xảy ra, vui lòng thử lại.", nil)
	}
}
