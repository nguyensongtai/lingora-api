package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/nguyensongtai/lingora-api/internal/api"
	"github.com/nguyensongtai/lingora-api/internal/httpx"
	"github.com/nguyensongtai/lingora-api/internal/user"
)

// service là những gì Handler cần ở tầng nghiệp vụ.
type service interface {
	Login(ctx context.Context, email, password string) (TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (TokenPair, error)
	Logout(ctx context.Context, refreshToken string) error
	CurrentUser(ctx context.Context, userID string) (user.User, error)
}

// Handler ánh xạ HTTP sang nghiệp vụ xác thực.
type Handler struct {
	service service
}

// NewHandler nhận service đã sẵn sàng dùng.
func NewHandler(svc service) *Handler {
	return &Handler{service: svc}
}

// Mount gắn route xác thực; authenticated bọc route cần token hợp lệ.
func (h *Handler) Mount(r chi.Router, authenticated func(http.Handler) http.Handler) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", h.login)
		r.Post("/refresh", h.refresh)
		r.Post("/logout", h.logout)
		r.With(authenticated).Get("/me", h.me)
	})
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
	switch {
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
