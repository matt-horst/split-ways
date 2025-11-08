-- name: CreateGroup :one
INSERT INTO groups (id, name, created_at, updated_at, owner)
VALUES (GEN_RANDOM_UUID(), $1, NOW(), NOW(), $2)
RETURNING *;

-- name: UpdateGroupName :one
UPDATE groups
SET name = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetGroupsByUser :many
SELECT groups.* FROM groups
INNER JOIN users_groups ON groups.id = users_groups.group_id
WHERE users_groups.user_id = $1;

-- name: GetUserGroup :one
SELECT * FROM users_groups
WHERE user_id = $1 AND group_id = $2;
