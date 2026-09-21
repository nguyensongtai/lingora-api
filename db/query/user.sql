-- name: CreateUser :one
INSERT INTO users (email, password_hash, display_name, role)
VALUES (
    sqlc.arg('email'),
    sqlc.arg('password_hash'),
    sqlc.arg('display_name'),
    sqlc.arg('role')
)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE lower(email) = lower(sqlc.arg('email')) AND deleted_at IS NULL;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;
