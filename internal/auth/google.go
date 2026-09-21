package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// googleTokenURL là endpoint đổi authorization code lấy token của Google.
const googleTokenURL = "https://oauth2.googleapis.com/token"

// googleIssuers là hai giá trị iss hợp lệ Google phát hành.
var googleIssuers = map[string]bool{
	"accounts.google.com":         true,
	"https://accounts.google.com": true,
}

// GoogleConfig là thông tin OAuth client. Bỏ trống nghĩa là tắt đăng nhập
// Google — không có giá trị mặc định nào dùng được ở đây.
type GoogleConfig struct {
	ClientID     string
	ClientSecret string
	// TokenURL để trống thì dùng endpoint thật của Google; test trỏ nó đi nơi khác.
	TokenURL string
}

// Configured cho biết có đủ thông tin để bật đăng nhập Google hay chưa.
func (c GoogleConfig) Configured() bool {
	return c.ClientID != "" && c.ClientSecret != ""
}

// GoogleIdentity là phần thông tin API cần từ tài khoản Google.
type GoogleIdentity struct {
	// Subject là id ổn định của tài khoản; email có thể đổi, subject thì không.
	Subject       string
	Email         string
	EmailVerified bool
	Name          string
}

// GoogleClient đổi authorization code lấy danh tính.
type GoogleClient struct {
	cfg  GoogleConfig
	http *http.Client
}

func NewGoogleClient(cfg GoogleConfig) *GoogleClient {
	return &GoogleClient{
		cfg:  cfg,
		http: &http.Client{Timeout: 10 * time.Second},
	}
}

type googleTokenResponse struct {
	IDToken string `json:"id_token"`
	Error   string `json:"error"`
}

type googleClaims struct {
	Issuer        string `json:"iss"`
	Audience      string `json:"aud"`
	Subject       string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	ExpiresAt     int64  `json:"exp"`
}

// Exchange đổi authorization code lấy danh tính Google.
//
// Không kiểm chữ ký của id_token: token này đến thẳng từ endpoint của Google
// qua TLS chứ không đi qua trình duyệt, nên chính kênh TLS đã là bằng chứng
// nguồn gốc. Vẫn kiểm iss và aud để một client secret bị dùng nhầm cho dự án
// khác không lọt qua.
func (c *GoogleClient) Exchange(ctx context.Context, code, redirectURI string) (GoogleIdentity, error) {
	endpoint := c.cfg.TokenURL
	if endpoint == "" {
		endpoint = googleTokenURL
	}

	form := url.Values{
		"code":          {code},
		"client_id":     {c.cfg.ClientID},
		"client_secret": {c.cfg.ClientSecret},
		"redirect_uri":  {redirectURI},
		"grant_type":    {"authorization_code"},
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return GoogleIdentity{}, fmt.Errorf("google exchange: %w", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := c.http.Do(request)
	if err != nil {
		return GoogleIdentity{}, fmt.Errorf("google exchange: %w", err)
	}
	defer response.Body.Close()

	var body googleTokenResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		return GoogleIdentity{}, fmt.Errorf("google exchange: decode response: %w", err)
	}
	if response.StatusCode != http.StatusOK || body.IDToken == "" {
		// Lý do cụ thể của Google chỉ hữu ích trong log; với client thì mọi
		// trường hợp đều là "code không dùng được".
		return GoogleIdentity{}, fmt.Errorf("google exchange: %s: %w", body.Error, ErrGoogleExchangeFailed)
	}

	claims, err := parseGoogleClaims(body.IDToken)
	if err != nil {
		return GoogleIdentity{}, err
	}
	if !googleIssuers[claims.Issuer] {
		return GoogleIdentity{}, fmt.Errorf("google exchange: iss %q: %w", claims.Issuer, ErrGoogleExchangeFailed)
	}
	if claims.Audience != c.cfg.ClientID {
		return GoogleIdentity{}, fmt.Errorf("google exchange: aud không khớp client id: %w", ErrGoogleExchangeFailed)
	}
	if claims.Subject == "" || claims.Email == "" {
		return GoogleIdentity{}, fmt.Errorf("google exchange: thiếu sub hoặc email: %w", ErrGoogleExchangeFailed)
	}

	return GoogleIdentity{
		Subject:       claims.Subject,
		Email:         claims.Email,
		EmailVerified: claims.EmailVerified,
		Name:          claims.Name,
	}, nil
}

// parseGoogleClaims đọc phần payload của id_token. Chỉ giải mã, không xác minh
// — xem chú thích của Exchange về lý do.
func parseGoogleClaims(idToken string) (googleClaims, error) {
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return googleClaims{}, fmt.Errorf("google exchange: id_token không đúng dạng JWT: %w", ErrGoogleExchangeFailed)
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return googleClaims{}, fmt.Errorf("google exchange: giải mã payload: %w", ErrGoogleExchangeFailed)
	}

	var claims googleClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return googleClaims{}, fmt.Errorf("google exchange: đọc claims: %w", ErrGoogleExchangeFailed)
	}
	return claims, nil
}
