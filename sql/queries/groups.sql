-- name: CreateGroup :one
INSERT INTO groups (id, name, created_at, updated_at)
VALUES (GEN_RANDOM_UUID(), $1, NOW(), NOW())
RETURNING *;

-- name: UpdateGroupName :one
UPDATE groups
SET name = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;
