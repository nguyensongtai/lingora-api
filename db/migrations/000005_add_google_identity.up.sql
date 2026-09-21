-- Tài khoản đăng nhập bằng Google không có mật khẩu, nên password_hash thôi bắt
-- buộc. Đổi lại phải có ràng buộc mới: mỗi tài khoản cần ít nhất một cách vào.
ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;

ALTER TABLE users DROP CONSTRAINT users_password_hash_not_blank;
ALTER TABLE users ADD CONSTRAINT users_password_hash_not_blank
    CHECK (password_hash IS NULL OR length(password_hash) > 0);

-- google_sub là id ổn định Google cấp cho mỗi tài khoản. Không dùng email làm
-- khoá: Google cho phép đổi email, còn sub thì không đổi.
ALTER TABLE users ADD COLUMN google_sub text;

ALTER TABLE users ADD CONSTRAINT users_google_sub_not_blank
    CHECK (google_sub IS NULL OR length(btrim(google_sub)) > 0);

ALTER TABLE users ADD CONSTRAINT users_has_credential
    CHECK (password_hash IS NOT NULL OR google_sub IS NOT NULL);

CREATE UNIQUE INDEX users_google_sub_key ON users (google_sub)
    WHERE google_sub IS NOT NULL AND deleted_at IS NULL;
