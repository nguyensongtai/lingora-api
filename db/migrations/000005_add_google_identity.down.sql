DROP INDEX users_google_sub_key;

ALTER TABLE users DROP CONSTRAINT users_has_credential;
ALTER TABLE users DROP CONSTRAINT users_google_sub_not_blank;
ALTER TABLE users DROP COLUMN google_sub;

-- Quay lại bắt buộc mật khẩu thì tài khoản chỉ có Google sẽ không hợp lệ; xoá
-- chúng đi là cách duy nhất để ràng buộc cũ áp lại được.
DELETE FROM users WHERE password_hash IS NULL;

ALTER TABLE users DROP CONSTRAINT users_password_hash_not_blank;
ALTER TABLE users ADD CONSTRAINT users_password_hash_not_blank
    CHECK (length(password_hash) > 0);
ALTER TABLE users ALTER COLUMN password_hash SET NOT NULL;
