-- name: CreateUserGroup :one
INSERT INTO users_groups (id, user_id, group_id)
VALUES (GEN_RANDOM_UUID(), $1, $2)
RETURNING *;
