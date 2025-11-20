-- name: CreateUserGroup :one
INSERT INTO users_groups (id, user_id, group_id)
VALUES (GEN_RANDOM_UUID(), $1, $2)
RETURNING *;

-- name: DeleteUserGroup :one
DELETE FROM users_groups
WHERE group_id = $1 AND user_id = $2
RETURNING *;
