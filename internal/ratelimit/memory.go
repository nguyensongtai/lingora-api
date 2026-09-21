package ratelimit

import (
	"context"
	"sync"
	"time"
)

// sweepInterval là khoảng cách tối thiểu giữa hai lần dọn khoá hết hạn. Dọn lười
// trong lúc phục vụ request thay vì chạy goroutine nền: không có vòng đời nào
// phải quản, và với số khoá của màn đăng nhập thì chi phí không đáng kể.
const sweepInterval = time.Minute

type window struct {
	count   int
	resetAt time.Time
}

// Memory đếm lượt trong bộ nhớ của một tiến trình.
//
// Đủ cho một instance, KHÔNG đủ khi chạy nhiều bản song song: mỗi bản có bộ đếm
// riêng nên hạn mức thực tế nhân lên theo số instance. Chạy trên Cloud Run có
// autoscale thì cần một kho dùng chung.
type Memory struct {
	mu        sync.Mutex
	windows   map[string]window
	lastSweep time.Time

	now func() time.Time
}

// NewMemory dựng limiter rỗng.
func NewMemory() *Memory {
	return &Memory{windows: make(map[string]window), now: time.Now}
}

// Allow ghi nhận một lượt theo cửa sổ cố định.
func (m *Memory) Allow(_ context.Context, key string, limit int, duration time.Duration) (Decision, error) {
	now := m.now()

	m.mu.Lock()
	defer m.mu.Unlock()
	m.sweep(now)

	current, exists := m.windows[key]
	if !exists || !now.Before(current.resetAt) {
		current = window{resetAt: now.Add(duration)}
	}

	if current.count >= limit {
		return Decision{RetryAfter: current.resetAt.Sub(now)}, nil
	}

	current.count++
	m.windows[key] = current
	return Decision{Allowed: true, Remaining: limit - current.count}, nil
}

// Reset xoá bộ đếm của một khoá.
func (m *Memory) Reset(_ context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.windows, key)
	return nil
}

// sweep bỏ khoá đã hết cửa sổ. Phải giữ khoá mu khi gọi.
func (m *Memory) sweep(now time.Time) {
	if now.Sub(m.lastSweep) < sweepInterval {
		return
	}
	m.lastSweep = now

	for key, value := range m.windows {
		if !now.Before(value.resetAt) {
			delete(m.windows, key)
		}
	}
}
