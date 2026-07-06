-- name: CreateUser :one
INSERT INTO users (google_id, email, name, avatar_url)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUserByGoogleID :one
SELECT * FROM users
WHERE google_id = $1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;