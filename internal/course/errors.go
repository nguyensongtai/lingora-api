package course

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

var (
	// ErrNotFound: khoá học không tồn tại hoặc đã bị xoá mềm.
	ErrNotFound = errors.New("course: not found")

	// ErrSlugTaken: slug đã được một khoá học chưa xoá khác sử dụng.
	ErrSlugTaken = errors.New("course: slug already taken")

	// ErrInvalidID: id truyền vào không phải UUID hợp lệ.
	ErrInvalidID = errors.New("course: invalid id")

	// ErrValidation: dữ liệu đầu vào không hợp lệ. Dùng errors.As với
	// *ValidationError để lấy chi tiết từng field.
	ErrValidation = errors.New("course: validation failed")
)

// ValidationError gom lỗi theo từng field để handler map thẳng sang
// details của body lỗi {code, message, details}.
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string {
	names := make([]string, 0, len(e.Fields))
	for name := range e.Fields {
		names = append(names, name)
	}
	sort.Strings(names)

	parts := make([]string, 0, len(names))
	for _, name := range names {
		parts = append(parts, fmt.Sprintf("%s: %s", name, e.Fields[name]))
	}
	return "course: validation failed (" + strings.Join(parts, "; ") + ")"
}

// Unwrap cho phép errors.Is(err, ErrValidation) nhận diện mọi lỗi validate.
func (e *ValidationError) Unwrap() error { return ErrValidation }

type validationBuilder struct {
	fields map[string]string
}

func (b *validationBuilder) add(field, message string) {
	if b.fields == nil {
		b.fields = make(map[string]string)
	}
	if _, exists := b.fields[field]; !exists {
		b.fields[field] = message
	}
}

// err trả về nil khi không có lỗi nào được ghi nhận.
func (b *validationBuilder) err() error {
	if len(b.fields) == 0 {
		return nil
	}
	return &ValidationError{Fields: b.fields}
}
