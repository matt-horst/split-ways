-- name: CreateGroup :one
INSERT INTO groups (id, name, created_at, updated_at, owner)
VALUES (GEN_RANDOM_UUID(), $1, NOW(), NOW(), $2)
RETURNING *;

-- name: UpdateGroupName :one
UPDATE groups
SET name = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;
