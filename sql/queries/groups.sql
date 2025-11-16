-- name: CreateGroup :one
INSERT INTO groups (name, owner)
VALUES ($1, $2)
RETURNING *;

-- name: UpdateGroupName :one
UPDATE groups
SET name = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetGroup :one
SELECT * FROM groups
WHERE id = $1;

-- name: GetGroupsByUser :many
SELECT groups.* FROM groups
INNER JOIN users_groups ON groups.id = users_groups.group_id
WHERE users_groups.user_id = $1;

-- name: GetUserGroup :one
SELECT * FROM users_groups
WHERE user_id = $1 AND group_id = $2;

-- name: GetUsersByGroup :many
SELECT users.* FROM users
INNER JOIN users_groups ON users.id = users_groups.user_id
WHERE users_groups.group_id = $1;

-- name: GetOtherUsersInGroup :many
SELECT users_groups.group_id AS group_id, users.* FROM users
INNER JOIN users_groups ON users.id = users_groups.user_id
WHERE group_id = $1 AND users.id != $2;
