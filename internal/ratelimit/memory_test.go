package ratelimit_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/nguyensongtai/lingora-api/internal/ratelimit"
)

const window = 15 * time.Minute

func TestAllowStopsAtTheLimit(t *testing.T) {
	t.Parallel()

	limiter := ratelimit.NewMemory()
	ctx := context.Background()

	for attempt := 1; attempt <= 3; attempt++ {
		decision, err := limiter.Allow(ctx, "login:an@example.com", 3, window)
		if err != nil {
			t.Fatalf("Allow() returned error: %v", err)
		}
		if !decision.Allowed {
			t.Fatalf("lượt %d bị chặn, want cho qua", attempt)
		}
		if decision.Remaining != 3-attempt {
			t.Errorf("lượt %d: Remaining = %d, want %d", attempt, decision.Remaining, 3-attempt)
		}
	}

	decision, err := limiter.Allow(ctx, "login:an@example.com", 3, window)
	if err != nil {
		t.Fatalf("Allow() returned error: %v", err)
	}
	if decision.Allowed {
		t.Fatal("lượt thứ tư được cho qua, want bị chặn")
	}
	if decision.RetryAfter <= 0 || decision.RetryAfter > window {
		t.Errorf("RetryAfter = %v, want trong khoảng (0, %v]", decision.RetryAfter, window)
	}
}

func TestKeysAreCountedSeparately(t *testing.T) {
	t.Parallel()

	limiter := ratelimit.NewMemory()
	ctx := context.Background()

	for range 3 {
		if _, err := limiter.Allow(ctx, "login:an@example.com", 3, window); err != nil {
			t.Fatalf("Allow() returned error: %v", err)
		}
	}

	// Người khác không bị vạ lây vì một tài khoản bị dò.
	other, err := limiter.Allow(ctx, "login:binh@example.com", 3, window)
	if err != nil {
		t.Fatalf("Allow() returned error: %v", err)
	}
	if !other.Allowed {
		t.Error("khoá khác bị chặn theo, want đếm riêng")
	}
}

func TestResetClearsTheCounter(t *testing.T) {
	t.Parallel()

	limiter := ratelimit.NewMemory()
	ctx := context.Background()

	for range 3 {
		if _, err := limiter.Allow(ctx, "login:an@example.com", 3, window); err != nil {
			t.Fatalf("Allow() returned error: %v", err)
		}
	}
	if err := limiter.Reset(ctx, "login:an@example.com"); err != nil {
		t.Fatalf("Reset() returned error: %v", err)
	}

	// Đăng nhập đúng xong thì lần gõ sai kế tiếp không được tính tiếp vào cửa
	// sổ cũ, nếu không một người hay quên sẽ tự khoá mình.
	decision, err := limiter.Allow(ctx, "login:an@example.com", 3, window)
	if err != nil {
		t.Fatalf("Allow() returned error: %v", err)
	}
	if !decision.Allowed {
		t.Error("sau Reset vẫn bị chặn, want đếm lại từ đầu")
	}
}

func TestAllowIsSafeUnderConcurrency(t *testing.T) {
	t.Parallel()

	limiter := ratelimit.NewMemory()
	ctx := context.Background()

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		allowed int
	)
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			decision, err := limiter.Allow(ctx, "login:an@example.com", 10, window)
			if err != nil {
				t.Errorf("Allow() returned error: %v", err)
				return
			}
			if decision.Allowed {
				mu.Lock()
				allowed++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	// Đúng 10 lượt lọt qua dù 50 goroutine cùng hỏi.
	if allowed != 10 {
		t.Errorf("cho qua %d lượt, want đúng 10", allowed)
	}
}

func TestRuleKeyNamespacesIdentifiers(t *testing.T) {
	t.Parallel()

	byIP := ratelimit.Rule{Name: "login-ip", Limit: 10, Window: window}
	byEmail := ratelimit.Rule{Name: "login-email", Limit: 5, Window: window}

	// Hai hạn mức khác nhau cùng một định danh không được đụng bộ đếm của nhau.
	if byIP.Key("x") == byEmail.Key("x") {
		t.Errorf("Key() trùng nhau: %q", byIP.Key("x"))
	}
}
