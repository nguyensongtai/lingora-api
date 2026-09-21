package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// incrementAndExpire tăng bộ đếm và đặt hạn cho khoá trong đúng một vòng tới
// Redis. Chạy nguyên khối nên hai instance cùng gọi không thể chen vào giữa
// INCR và EXPIRE để tạo ra một khoá sống mãi.
//
// Trả về [số lượt hiện tại, số mili giây còn lại của cửa sổ].
var incrementAndExpire = redis.NewScript(`
local count = redis.call('INCR', KEYS[1])
if count == 1 then
  redis.call('PEXPIRE', KEYS[1], ARGV[1])
end
return {count, redis.call('PTTL', KEYS[1])}
`)

// Redis đếm lượt trên một kho dùng chung, nên hạn mức đúng kể cả khi API chạy
// nhiều instance song song.
type Redis struct {
	client redis.UniversalClient
	prefix string
}

// NewRedis dựng limiter trên một client đã sẵn sàng. prefix tách không gian
// khoá của tính năng này khỏi mọi thứ khác dùng chung Redis.
func NewRedis(client redis.UniversalClient, prefix string) *Redis {
	return &Redis{client: client, prefix: prefix}
}

// Allow ghi nhận một lượt theo cửa sổ cố định.
func (r *Redis) Allow(ctx context.Context, key string, limit int, duration time.Duration) (Decision, error) {
	result, err := incrementAndExpire.Run(
		ctx, r.client, []string{r.key(key)}, duration.Milliseconds(),
	).Int64Slice()
	if err != nil {
		return Decision{}, fmt.Errorf("ratelimit: increment %q: %w", key, err)
	}
	if len(result) != 2 {
		return Decision{}, fmt.Errorf("ratelimit: increment %q: kết quả %d phần tử, cần 2", key, len(result))
	}

	count, ttl := result[0], result[1]
	if count > int64(limit) {
		retryAfter := time.Duration(ttl) * time.Millisecond
		if ttl < 0 {
			// PTTL âm nghĩa là khoá không có hạn — không nên xảy ra, nhưng thà
			// hẹn lại đúng một cửa sổ còn hơn trả thời gian âm.
			retryAfter = duration
		}
		return Decision{RetryAfter: retryAfter}, nil
	}

	return Decision{Allowed: true, Remaining: limit - int(count)}, nil
}

// Reset xoá bộ đếm của một khoá.
func (r *Redis) Reset(ctx context.Context, key string) error {
	if err := r.client.Del(ctx, r.key(key)).Err(); err != nil {
		return fmt.Errorf("ratelimit: reset %q: %w", key, err)
	}
	return nil
}

func (r *Redis) key(key string) string {
	return r.prefix + key
}
