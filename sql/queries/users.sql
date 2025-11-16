-- name: CreateUser :one
INSERT INTO users (username, hashed_password)
VALUES ($1, $2)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = $1;

-- name: UpdatePassword :one
UPDATE users
SET hashed_password = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;
