// Package auth xác thực JWT và kiểm tra quyền cho các route ghi.
package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"github.com/nguyensongtai/lingora-api/internal/httpx"
)

// RoleAdmin là quyền bắt buộc cho mọi thao tác ghi nội dung.
const RoleAdmin = "admin"

// minSecretLen giữ HS256 ở mức khoá đủ dài.
const minSecretLen = 32

// Claims là payload token mà API quan tâm.
type Claims struct {
	Subject string
	Role    string
}

type tokenClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

type contextKey struct{}

// Verifier kiểm tra chữ ký HS256 của access token.
type Verifier struct {
	secret []byte
}

// NewVerifier từ chối secret yếu ngay lúc boot thay vì lúc có request đầu tiên.
func NewVerifier(secret string) (*Verifier, error) {
	if len(secret) < minSecretLen {
		return nil, fmt.Errorf("%w (got %d)", ErrWeakSecret, len(secret))
	}
	return &Verifier{secret: []byte(secret)}, nil
}

// RequireAuthenticated trả về middleware chỉ bắt buộc token hợp lệ, không xét role.
func (v *Verifier) RequireAuthenticated() func(http.Handler) http.Handler {
	return v.middleware("")
}

// RequireRole trả về middleware bắt buộc token hợp lệ và đúng role.
func (v *Verifier) RequireRole(role string) func(http.Handler) http.Handler {
	return v.middleware(role)
}

// OptionalAuth đọc token nếu có và gắn claims vào context, nhưng không chặn ai
// cả. Dùng cho route công khai mà nội dung trả về vẫn phụ thuộc người đọc —
// ví dụ khoá học: khách thấy khoá đã xuất bản, admin thấy cả bản nháp.
//
// Token hỏng hay hết hạn bị coi như không có, chứ không trả 401: route này vốn
// mở cho khách, nên một token cũ trong tab bỏ quên không được phép biến nó
// thành lỗi. Hệ quả là token xấu không bao giờ mở thêm được gì.
func (v *Verifier) OptionalAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw, err := bearerToken(r)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			claims, err := v.parse(raw)
			if err != nil {
				slog.WarnContext(r.Context(), "ignore access token on public route", slog.Any("error", err))
				next.ServeHTTP(w, r)
				return
			}

			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), contextKey{}, claims)))
		})
	}
}

// middleware với role rỗng nghĩa là chấp nhận mọi role.
func (v *Verifier) middleware(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw, err := bearerToken(r)
			if err != nil {
				httpx.Error(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "Thiếu access token hợp lệ.", nil)
				return
			}

			claims, err := v.parse(raw)
			if err != nil {
				// Lý do cụ thể chỉ ghi log, không trả cho client.
				slog.WarnContext(r.Context(), "reject access token", slog.Any("error", err))
				httpx.Error(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "Access token không hợp lệ hoặc đã hết hạn.", nil)
				return
			}

			if role != "" && claims.Role != role {
				httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "Tài khoản không có quyền thực hiện thao tác này.", nil)
				return
			}

			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), contextKey{}, claims)))
		})
	}
}

// ClaimsFrom lấy claims mà middleware đã gắn vào context.
func ClaimsFrom(ctx context.Context) (Claims, bool) {
	claims, ok := ctx.Value(contextKey{}).(Claims)
	return claims, ok
}

func (v *Verifier) parse(raw string) (Claims, error) {
	var parsed tokenClaims

	// Chỉ chấp nhận HS256: khoá alg mở là lỗ hổng kinh điển của JWT.
	if _, err := jwt.ParseWithClaims(raw, &parsed, func(*jwt.Token) (any, error) {
		return v.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired()); err != nil {
		return Claims{}, fmt.Errorf("parse token: %w", err)
	}

	return Claims{Subject: parsed.Subject, Role: parsed.Role}, nil
}

func bearerToken(r *http.Request) (string, error) {
	header := r.Header.Get("Authorization")
	scheme, token, found := strings.Cut(header, " ")
	if !found || !strings.EqualFold(scheme, "bearer") || strings.TrimSpace(token) == "" {
		return "", errors.New("auth: missing bearer token")
	}
	return strings.TrimSpace(token), nil
}
