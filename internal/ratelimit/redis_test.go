package ratelimit_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/nguyensongtai/lingora-api/internal/ratelimit"
)

// redisLimiter dựng limiter trên Redis thật. Không có REDIS_URL thì bỏ qua:
// test này cần một Redis đang chạy, và bắt buộc nó sẽ làm gãy máy của người
// chưa bật docker compose.
func redisLimiter(t *testing.T) *ratelimit.Redis {
	t.Helper()

	url := os.Getenv("REDIS_URL")
	if url == "" {
		t.Skip("REDIS_URL chưa đặt, bỏ qua test cần Redis thật")
	}

	options, err := redis.ParseURL(url)
	if err != nil {
		t.Fatalf("parse REDIS_URL: %v", err)
	}

	client := redis.NewClient(options)
	t.Cleanup(func() { client.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("không kết nối được Redis: %v", err)
	}

	// Mỗi lần chạy một prefix riêng để hai test không giẫm lên nhau.
	return ratelimit.NewRedis(client, "lingora:test:"+t.Name()+":")
}

func TestRedisAllowStopsAtTheLimit(t *testing.T) {
	t.Parallel()

	limiter := redisLimiter(t)
	ctx := context.Background()
	key := "login:an@example.com"
	t.Cleanup(func() { _ = limiter.Reset(ctx, key) })

	for attempt := 1; attempt <= 3; attempt++ {
		decision, err := limiter.Allow(ctx, key, 3, window)
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

	decision, err := limiter.Allow(ctx, key, 3, window)
	if err != nil {
		t.Fatalf("Allow() returned error: %v", err)
	}
	if decision.Allowed {
		t.Fatal("lượt thứ tư được cho qua, want bị chặn")
	}
	// PTTL phải trả về phần còn lại của cửa sổ, không phải trọn cửa sổ.
	if decision.RetryAfter <= 0 || decision.RetryAfter > window {
		t.Errorf("RetryAfter = %v, want trong khoảng (0, %v]", decision.RetryAfter, window)
	}
}

func TestRedisWindowExpires(t *testing.T) {
	t.Parallel()

	limiter := redisLimiter(t)
	ctx := context.Background()
	key := "short:an@example.com"
	t.Cleanup(func() { _ = limiter.Reset(ctx, key) })

	short := 300 * time.Millisecond
	if _, err := limiter.Allow(ctx, key, 1, short); err != nil {
		t.Fatalf("Allow() returned error: %v", err)
	}
	blocked, err := limiter.Allow(ctx, key, 1, short)
	if err != nil {
		t.Fatalf("Allow() returned error: %v", err)
	}
	if blocked.Allowed {
		t.Fatal("lượt thứ hai được cho qua, want bị chặn")
	}

	// Hạn của khoá phải do PEXPIRE đặt ở lượt đầu; hết cửa sổ là đếm lại.
	time.Sleep(400 * time.Millisecond)

	after, err := limiter.Allow(ctx, key, 1, short)
	if err != nil {
		t.Fatalf("Allow() returned error: %v", err)
	}
	if !after.Allowed {
		t.Error("sau khi hết cửa sổ vẫn bị chặn, want đếm lại từ đầu")
	}
}

func TestRedisResetClearsTheCounter(t *testing.T) {
	t.Parallel()

	limiter := redisLimiter(t)
	ctx := context.Background()
	key := "reset:an@example.com"
	t.Cleanup(func() { _ = limiter.Reset(ctx, key) })

	if _, err := limiter.Allow(ctx, key, 1, window); err != nil {
		t.Fatalf("Allow() returned error: %v", err)
	}
	if err := limiter.Reset(ctx, key); err != nil {
		t.Fatalf("Reset() returned error: %v", err)
	}

	decision, err := limiter.Allow(ctx, key, 1, window)
	if err != nil {
		t.Fatalf("Allow() returned error: %v", err)
	}
	if !decision.Allowed {
		t.Error("sau Reset vẫn bị chặn, want đếm lại từ đầu")
	}
}
