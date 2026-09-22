// Package httpx chứa phần dùng chung của tầng HTTP: body lỗi thống nhất,
// đọc/ghi JSON và middleware CORS.
package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

// maxBodyBytes chặn body quá lớn trước khi nó kịp vào bộ nhớ.
const maxBodyBytes = 1 << 20 // 1 MiB

// ErrMalformedBody: body không phải JSON hợp lệ hoặc sai kiểu.
var ErrMalformedBody = errors.New("httpx: malformed request body")

// Mã lỗi dùng chung cho toàn bộ API.
const (
	CodeValidation       = "validation_error"
	CodeMalformed        = "malformed_body"
	CodeNotFound         = "not_found"
	CodeConflict         = "conflict"
	CodeUnauthorized     = "unauthorized"
	CodeForbidden        = "forbidden"
	CodeRateLimited      = "rate_limited"
	CodeMethodNotAllowed = "method_not_allowed"
	CodeInternal         = "internal_error"
)

// ErrorResponse là body lỗi thống nhất của toàn bộ API.
type ErrorResponse struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// JSON ghi một response thành công.
func JSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if body == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(body); err != nil {
		// Header đã gửi đi rồi nên chỉ còn cách ghi log.
		slog.Error("encode response body", slog.Any("error", err))
	}
}

// NoContent trả về 204.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Error ghi body lỗi thống nhất {code, message, details}.
func Error(w http.ResponseWriter, status int, code, message string, details map[string]string) {
	JSON(w, status, ErrorResponse{Code: code, Message: message, Details: details})
}

// DecodeJSON đọc đúng một object JSON vào dst. Field lạ bị từ chối để lỗi
// chính tả phía client không âm thầm bị bỏ qua.
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	if contentType := r.Header.Get("Content-Type"); contentType != "" {
		if mediaType, _, _ := strings.Cut(contentType, ";"); strings.TrimSpace(mediaType) != "application/json" {
			return fmt.Errorf("%w: content type %q is not application/json", ErrMalformedBody, mediaType)
		}
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return fmt.Errorf("%w: %s", ErrMalformedBody, err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return fmt.Errorf("%w: body must contain a single JSON object", ErrMalformedBody)
	}
	return nil
}

// Nullable phân biệt ba trạng thái của một field trong PATCH body: vắng mặt,
// đặt bằng null, và đặt bằng một giá trị.
type Nullable[T any] struct {
	Set   bool
	Value *T
}

// UnmarshalJSON đánh dấu Set ngay khi key có mặt trong body.
func (n *Nullable[T]) UnmarshalJSON(data []byte) error {
	n.Set = true
	if string(data) == "null" {
		n.Value = nil
		return nil
	}
	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	n.Value = &value
	return nil
}

// IsNull cho biết client gửi đúng giá trị null (khác với không gửi field).
func (n Nullable[T]) IsNull() bool { return n.Set && n.Value == nil }
