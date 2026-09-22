// Package testdb mở kết nối tới một PostgreSQL thật cho integration test.
//
// Mỗi test chạy trong một transaction rồi rollback, nên chúng không thấy nhau
// và không để lại gì trong database — kể cả khi chạy trên chính database dev.
package testdb

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	poolOnce sync.Once
	pool     *pgxpool.Pool
	poolErr  error
)

// URL là biến môi trường trỏ tới database dùng cho test.
const URL = "TEST_DATABASE_URL"

// New trả về một transaction đã sẵn sàng, tự rollback khi test kết thúc.
//
// Thiếu TEST_DATABASE_URL thì bỏ qua test: `go test ./...` phải chạy được trên
// máy chưa bật docker compose.
func New(t *testing.T) pgx.Tx {
	t.Helper()

	dsn := os.Getenv(URL)
	if dsn == "" {
		t.Skipf("%s chưa đặt, bỏ qua test cần PostgreSQL thật", URL)
	}

	poolOnce.Do(func() { pool, poolErr = connect(dsn) })
	if poolErr != nil {
		t.Skipf("không kết nối được PostgreSQL: %v", poolErr)
	}

	tx, err := pool.Begin(context.Background())
	if err != nil {
		t.Fatalf("begin transaction: %v", err)
	}
	t.Cleanup(func() {
		// Rollback luôn, kể cả khi test thành công: dữ liệu của test không được
		// sống sót sang test sau.
		_ = tx.Rollback(context.Background())
	})
	return tx
}

func connect(dsn string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	created, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	if err := created.Ping(ctx); err != nil {
		created.Close()
		return nil, err
	}
	if err := ensureSchema(ctx, created); err != nil {
		created.Close()
		return nil, err
	}
	return created, nil
}

// schemaLockKey là khoá advisory lock cho việc dựng schema. Giá trị tuỳ ý,
// chỉ cần mọi tiến trình test dùng chung một số.
const schemaLockKey int64 = 0x1_0206_2A11

// ensureSchema chạy các migration lên nếu database còn trống. Database đã có
// schema thì không đụng tới — chạy lại migration trên database dev sẽ hỏng.
//
// `go test ./...` chạy mỗi package trong MỘT TIẾN TRÌNH RIÊNG, tất cả trỏ vào
// cùng một database, nên sync.Once ở trên không đủ. Hai lớp chống đua:
//
//   - advisory lock để đúng một tiến trình dựng schema, những tiến trình khác
//     chờ tới lượt rồi mới kiểm tra — thay vì kiểm xong mới biết là vừa đua;
//   - toàn bộ migration nằm trong một transaction, nên không tiến trình nào
//     nhìn thấy schema dựng dở. Trước đây mốc kiểm là bảng `courses` do
//     migration đầu tiên tạo, nên một tiến trình có thể thấy nó rồi bỏ qua
//     trong khi các migration sau chưa chạy — CI đỏ vì thiếu cột `position`
//     và `google_sub`.
func ensureSchema(ctx context.Context, pool *pgxpool.Pool) error {
	// Lock nằm trên session nên phải giữ nguyên một connection.
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, "SELECT pg_advisory_lock($1)", schemaLockKey); err != nil {
		return err
	}
	defer func() {
		_, _ = conn.Exec(context.WithoutCancel(ctx), "SELECT pg_advisory_unlock($1)", schemaLockKey)
	}()

	var exists *string
	if err := conn.QueryRow(ctx, "SELECT to_regclass('public.courses')::text").Scan(&exists); err != nil {
		return err
	}
	if exists != nil {
		return nil
	}

	files, err := migrationFiles()
	if err != nil {
		return err
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()

	for _, file := range files {
		statements, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, string(statements)); err != nil {
			return fmt.Errorf("apply %s: %w", filepath.Base(file), err)
		}
	}
	return tx.Commit(ctx)
}

// migrationFiles trả về các file .up.sql theo đúng thứ tự migration.
func migrationFiles() ([]string, error) {
	dir, err := migrationsDir()
	if err != nil {
		return nil, err
	}
	files, err := filepath.Glob(filepath.Join(dir, "*.up.sql"))
	if err != nil {
		return nil, err
	}
	// Tên file có tiền tố số thứ tự nên sắp theo tên là đúng thứ tự migration.
	sort.Strings(files)
	return files, nil
}

// migrationsDir đi ngược lên từ thư mục của test cho tới khi thấy db/migrations.
// Test chạy với working directory là thư mục package, nên không thể viết cứng
// một đường dẫn tương đối.
func migrationsDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		candidate := filepath.Join(dir, "db", "migrations")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}

// Attempt chạy một thao tác được dự đoán sẽ lỗi, bên trong một savepoint.
//
// PostgreSQL huỷ cả transaction ngay khi một câu lệnh lỗi, nên không có
// savepoint thì một test kiểm ràng buộc sẽ không dùng tiếp được transaction của
// chính nó. Rollback về savepoint đưa transaction trở lại dùng được.
func Attempt(t *testing.T, tx pgx.Tx, fn func(pgx.Tx) error) error {
	t.Helper()

	ctx := context.Background()
	nested, err := tx.Begin(ctx)
	if err != nil {
		t.Fatalf("begin savepoint: %v", err)
	}

	attemptErr := fn(nested)
	if rollbackErr := nested.Rollback(ctx); rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
		t.Fatalf("rollback savepoint: %v", rollbackErr)
	}
	return attemptErr
}
