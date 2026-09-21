package vocabulary

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

var (
	// ErrNotFound: từ không tồn tại hoặc đã bị xoá mềm.
	ErrNotFound = errors.New("vocabulary: entry not found")

	// ErrWordTaken: từ này đã có trong cùng bài học.
	ErrWordTaken = errors.New("vocabulary: word already in this lesson")

	// ErrLessonNotFound: bài học không tồn tại hoặc đã bị xoá mềm.
	ErrLessonNotFound = errors.New("vocabulary: lesson not found")

	// ErrNotUnlocked: người học chưa hoàn thành bài chứa từ này nên chưa ôn được.
	ErrNotUnlocked = errors.New("vocabulary: entry not unlocked for this user")

	// ErrInvalidID: id truyền vào không phải UUID hợp lệ.
	ErrInvalidID = errors.New("vocabulary: invalid id")

	// ErrValidation: dữ liệu đầu vào không hợp lệ. Dùng errors.As với
	// *ValidationError để lấy chi tiết từng field.
	ErrValidation = errors.New("vocabulary: validation failed")
)

// ValidationError gom lỗi theo từng field để handler map thẳng sang details.
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
	return "vocabulary: validation failed (" + strings.Join(parts, "; ") + ")"
}

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

func (b *validationBuilder) err() error {
	if len(b.fields) == 0 {
		return nil
	}
	return &ValidationError{Fields: b.fields}
}
