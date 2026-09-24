-- Mục tiêu XP mỗi ngày, người học tự chọn ở trang tài khoản. Trước đây là hằng
-- số 50 trong code. Chỉ ba mức, như người dùng đã chốt: một ô nhập tự do thì
-- "3 XP" hay "5000 XP" đều hợp lệ và thanh mục tiêu mất nghĩa.
ALTER TABLE users ADD COLUMN daily_goal_xp integer NOT NULL DEFAULT 50;

ALTER TABLE users ADD CONSTRAINT users_daily_goal_xp_allowed
    CHECK (daily_goal_xp IN (20, 50, 100));
