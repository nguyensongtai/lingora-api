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
