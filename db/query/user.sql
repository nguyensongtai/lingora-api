-- name: CreateUser :one
INSERT INTO users (email, password_hash, display_name, role, google_sub)
VALUES (
    sqlc.arg('email'),
    sqlc.narg('password_hash'),
    sqlc.arg('display_name'),
    sqlc.arg('role'),
    sqlc.narg('google_sub')
)
RETURNING *;

-- name: GetUserByGoogleSub :one
SELECT * FROM users
WHERE google_sub = sqlc.arg('google_sub') AND deleted_at IS NULL;

-- name: LinkGoogleSub :one
-- Gắn tài khoản Google vào một tài khoản email đã có. Chỉ đổi khi chưa gắn với
-- ai, để hai lần gọi song song không ghi đè lên nhau.
UPDATE users
SET google_sub = sqlc.arg('google_sub'), updated_at = now()
WHERE id = sqlc.arg('id') AND google_sub IS NULL AND deleted_at IS NULL
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE lower(email) = lower(sqlc.arg('email')) AND deleted_at IS NULL;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: UpdateUserProfile :one
-- Partial update: NULL nghĩa là giữ nguyên.
UPDATE users
SET display_name  = COALESCE(sqlc.narg('display_name'), display_name),
    daily_goal_xp = COALESCE(sqlc.narg('daily_goal_xp'), daily_goal_xp),
    updated_at    = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING *;

-- name: UpdateUserPassword :execrows
-- Chỉ đổi được mật khẩu của tài khoản vốn đã có mật khẩu: tài khoản chỉ có
-- Google đặt mật khẩu lần đầu là một luồng khác, chưa làm.
UPDATE users
SET password_hash = sqlc.arg('password_hash'), updated_at = now()
WHERE id = sqlc.arg('id') AND password_hash IS NOT NULL AND deleted_at IS NULL;
