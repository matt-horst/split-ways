-- name: CreateExpense :one
INSERT INTO expenses (id, created_at, updated_at, group_id, created_by, paid_by, description)
VALUES (
    GEN_RANDOM_UUID(),
    NOW(),
    NOW(),
    $1,
    $2,
    $3,
    $4
) RETURNING *;

-- name: UpdateExpense :one
UPDATE expenses
SET paid_by = $2, description = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteExpense :one
DELETE FROM expenses
WHERE id = $1
RETURNING *;

-- name: GetExpensesByGroup :many
SELECT * FROM expenses
WHERE group_id = $1;

-- name: CreateSplit :one
INSERT INTO splits (id, created_at, updated_at, expense_id, user_id, amount)
VALUES (
    GEN_RANDOM_UUID(),
    NOW(),
    NOW(),
    $1,
    $2,
    $3
) RETURNING *;

-- name: DeleteSplit :one
DELETE FROM splits
WHERE id = $1
RETURNING *;

