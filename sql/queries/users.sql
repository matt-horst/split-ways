-- name: CreateUser :one
INSERT INTO users (id, username, hashed_password, created_at, updated_at)
VALUES (GEN_RANDOM_UUID(), $1, $2, NOW(), NOW())
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE id = $1;

-- name: UpdatePassword :one
UPDATE users
SET hashed_password = $2
WHERE id = $1
RETURNING *;
