-- name: CreateChirp :one
INSERT INTO chirps (
  id, created_at, updated_at, body, user_id
) VALUES (
  $1, $2, $3, $4, $5
  ) RETURNING *;

-- name: GetChirps :many
SELECT id, created_at, updated_at, body, user_id
FROM chirps
ORDER BY created_at;

-- name: GetChirpByID :one
SELECT id, created_at, updated_at, body, user_id
FROM chirps
WHERE id = $1;

-- name: DeleteChirp :one
DELETE FROM chirps WHERE id = $1 AND user_id = $2 RETURNING *;

-- name: GetChirpsByUserID :many
SELECT id, created_at, updated_at, body, user_id
FROM chirps
WHERE user_id = $1;
