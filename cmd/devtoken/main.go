// Command devtoken in ra một access token HS256 để gọi thử các endpoint admin
// khi chưa có domain auth thật. Chỉ dùng cho môi trường dev.
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	subject := flag.String("sub", "dev-admin", "giá trị claim sub")
	role := flag.String("role", "admin", "giá trị claim role")
	ttl := flag.Duration("ttl", time.Hour, "thời hạn token")
	flag.Parse()

	if err := run(*subject, *role, *ttl); err != nil {
		fmt.Fprintln(os.Stderr, "devtoken:", err)
		os.Exit(1)
	}
}

func run(subject, role string, ttl time.Duration) error {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return fmt.Errorf("JWT_SECRET chưa được đặt")
	}

	now := time.Now()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  subject,
		"role": role,
		"iat":  now.Unix(),
		"exp":  now.Add(ttl).Unix(),
	}).SignedString([]byte(secret))
	if err != nil {
		return fmt.Errorf("ký token: %w", err)
	}

	fmt.Println(token)
	return nil
}
