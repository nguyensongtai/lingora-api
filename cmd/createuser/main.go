// Command createuser tạo một tài khoản mới. Mật khẩu đọc từ stdin để không nằm
// lại trong lịch sử shell hay trong danh sách tiến trình.
//
//	read -rs PASSWORD && printf '%s' "$PASSWORD" | go run ./cmd/createuser -email admin@lingora.vn -role admin
package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/nguyensongtai/lingora-api/internal/auth"
	"github.com/nguyensongtai/lingora-api/internal/config"
	"github.com/nguyensongtai/lingora-api/internal/platform/postgres"
	"github.com/nguyensongtai/lingora-api/internal/user"
)

func main() {
	email := flag.String("email", "", "email đăng nhập")
	name := flag.String("name", "", "tên hiển thị (mặc định lấy phần trước @ của email)")
	role := flag.String("role", string(user.RoleStudent), "quyền: admin hoặc student")
	flag.Parse()

	if err := run(*email, *name, *role); err != nil {
		fmt.Fprintln(os.Stderr, "createuser:", err)
		os.Exit(1)
	}
}

func run(email, name, role string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return errors.New("thiếu -email")
	}
	if !user.Role(role).Valid() {
		return fmt.Errorf("role %q không hợp lệ, phải là admin hoặc student", role)
	}
	if name == "" {
		name, _, _ = strings.Cut(email, "@")
	}

	password, err := readPassword(os.Stdin)
	if err != nil {
		return err
	}
	if utf8.RuneCountInString(password) < auth.MinPasswordLen {
		return fmt.Errorf("mật khẩu phải dài tối thiểu %d ký tự", auth.MinPasswordLen)
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("đọc cấu hình: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("kết nối database: %w", err)
	}
	defer pool.Close()

	created, err := user.NewRepo(pool).Create(ctx, user.CreateParams{
		Email:        email,
		PasswordHash: &hash,
		DisplayName:  name,
		Role:         user.Role(role),
	})
	if err != nil {
		if errors.Is(err, user.ErrEmailTaken) {
			return fmt.Errorf("email %s đã được dùng", email)
		}
		return err
	}

	fmt.Printf("đã tạo %s (%s) với quyền %s\n", created.Email, created.ID, created.Role)
	return nil
}

func readPassword(in io.Reader) (string, error) {
	reader := bufio.NewReader(in)

	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("đọc mật khẩu từ stdin: %w", err)
	}

	password := strings.TrimRight(line, "\r\n")
	if password == "" {
		return "", errors.New("mật khẩu rỗng: truyền qua stdin, ví dụ `read -rs PW && printf '%s' \"$PW\" | ...`")
	}
	return password, nil
}
