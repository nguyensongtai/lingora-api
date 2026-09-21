package auth

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// bcryptCost 12 là cân bằng giữa chi phí tấn công và độ trễ đăng nhập.
const bcryptCost = 12

// MinPasswordLen là độ dài tối thiểu của mật khẩu mới.
const MinPasswordLen = 12

// dummyHash dùng để so sánh giả khi email không tồn tại, giữ thời gian phản hồi
// của hai nhánh tương đương nhau nên không dò được email nào có thật.
var dummyHash = []byte("$2a$12$eImiTXuWVxfM37uY4JANjQ.cGMHTRPRnEJLWdaNfJDCEGZkDzJHXe")

// HashPassword băm mật khẩu bằng bcrypt.
func HashPassword(plain string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hashed), nil
}

// VerifyPassword so mật khẩu với hash; sai mật khẩu trả về ErrInvalidCredentials.
func VerifyPassword(hash, plain string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidCredentials
		}
		return fmt.Errorf("compare password: %w", err)
	}
	return nil
}

// burnPasswordTime chạy bcrypt trên hash giả để nhánh "email không tồn tại" tốn
// thời gian tương đương nhánh "sai mật khẩu".
func burnPasswordTime(plain string) {
	_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(plain))
}
