-- name: CreatePayment :one
INSERT INTO payments (id, created_at, updated_at, group_id, created_by, paid_by, paid_to, amount)
VALUES (
    GEN_RANDOM_UUID(),
    NOW(),
    NOW(),
    $1,
    $2,
    $3,
    $4,
    $5
) RETURNING *;

-- name: UpdatePayment :one
UPDATE payments
SET amount = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeletePayment :one
DELETE FROM payments
WHERE id = $1
RETURNING *;
