-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (
  token,
  created_at, 
  updated_at, 
  user_id,
  expires_at,
  revoked_at
) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;
-- name: GetUserFromRefreshToken :one
SELECT users.* FROM users
INNER JOIN refresh_tokens ON users.id = refresh_tokens.user_id
WHERE token = $1
AND revoked_at IS NULL
AND expires_at > NOW();
-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens SET
  revoked_at = $1,
  updated_at = $2
WHERE token = $3;
