package api_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/nguyensongtai/lingora-api/internal/api"
	"github.com/nguyensongtai/lingora-api/internal/httpx"
)

// httpx.ErrorResponse dùng chung cho mọi domain nên không sinh từ spec được.
// Test này khoá nó lại đúng bằng schema Error trong openapi.yaml.
func TestErrorResponseMatchesSpecShape(t *testing.T) {
	t.Parallel()

	raw, err := json.Marshal(httpx.ErrorResponse{
		Code:    httpx.CodeValidation,
		Message: "Dữ liệu không hợp lệ.",
		Details: map[string]string{"slug": "không được rỗng"},
	})
	if err != nil {
		t.Fatalf("marshal httpx.ErrorResponse: %v", err)
	}

	var fromSpec api.Error
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&fromSpec); err != nil {
		t.Fatalf("body lỗi thật không khớp schema Error của spec: %v (body: %s)", err, raw)
	}

	if fromSpec.Code != api.ErrorCodeValidationError {
		t.Errorf("code = %q, want %q", fromSpec.Code, api.ErrorCodeValidationError)
	}
	if fromSpec.Details == nil || (*fromSpec.Details)["slug"] == "" {
		t.Errorf("details = %v, want the per-field entry to survive the round trip", fromSpec.Details)
	}
}

// Mọi mã lỗi httpx phát ra phải nằm trong enum của spec, nếu không client sinh
// type từ spec sẽ gặp giá trị nó không biết.
func TestEveryErrorCodeIsInTheSpecEnum(t *testing.T) {
	t.Parallel()

	codes := []string{
		httpx.CodeValidation,
		httpx.CodeMalformed,
		httpx.CodeNotFound,
		httpx.CodeConflict,
		httpx.CodeUnauthorized,
		httpx.CodeForbidden,
		httpx.CodeMethodNotAllowed,
		httpx.CodeRateLimited,
		httpx.CodeInternal,
	}

	for _, code := range codes {
		if !api.ErrorCode(code).Valid() {
			t.Errorf("mã lỗi %q không có trong enum Error.code của openapi.yaml", code)
		}
	}
}

// Enum của domain và enum của spec phải cùng tập giá trị.
func TestDomainEnumsMatchSpecEnums(t *testing.T) {
	t.Parallel()

	for _, level := range []string{"A1", "A2", "B1", "B2", "C1", "C2"} {
		if !api.CourseLevel(level).Valid() {
			t.Errorf("level %q có trong migration nhưng không có trong spec", level)
		}
	}
	for _, status := range []string{"draft", "published"} {
		if !api.CourseStatus(status).Valid() {
			t.Errorf("status %q có trong migration nhưng không có trong spec", status)
		}
	}
	for _, role := range []string{"admin", "student"} {
		if !api.UserRole(role).Valid() {
			t.Errorf("role %q có trong migration nhưng không có trong spec", role)
		}
	}
}
