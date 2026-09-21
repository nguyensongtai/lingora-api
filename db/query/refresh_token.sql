-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
VALUES (sqlc.arg('user_id'), sqlc.arg('token_hash'), sqlc.arg('expires_at'))
RETURNING *;

-- name: GetActiveRefreshToken :one
-- Chỉ trả về token còn hiệu lực; token hết hạn hoặc đã thu hồi coi như không có.
SELECT * FROM refresh_tokens
WHERE token_hash = sqlc.arg('token_hash')
  AND revoked_at IS NULL
  AND expires_at > now();

-- name: RevokeRefreshToken :execrows
UPDATE refresh_tokens
SET revoked_at = now()
WHERE token_hash = sqlc.arg('token_hash') AND revoked_at IS NULL;

-- name: RevokeAllUserRefreshTokens :execrows
UPDATE refresh_tokens
SET revoked_at = now()
WHERE user_id = sqlc.arg('user_id') AND revoked_at IS NULL;

-- name: DeleteExpiredRefreshTokens :execrows
DELETE FROM refresh_tokens
WHERE expires_at <= now() OR revoked_at IS NOT NULL;
