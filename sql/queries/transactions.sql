-- name: CreateExpense :one
INSERT INTO expenses (transaction_id, paid_by, description, amount)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateExpense :one
UPDATE expenses
SET paid_by = $2, description = $3, amount = $4
WHERE id = $1
RETURNING *;

-- name: GetExpensesByGroup :many
SELECT expenses.* FROM expenses
INNER JOIN transactions ON expenses.transaction_id = transactions.id
WHERE transactions.group_id = $1
ORDER BY transactions.updated_at;

-- name: CreateDebt :one
INSERT INTO debts (expense_id, owed_to, owed_by, amount)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateDebt :one
UPDATE debts
SET amount = $2
WHERE id = $1
RETURNING *;

-- name: DeleteDebtsByExpense :exec
DELETE FROM debts
WHERE expense_id = $1;

-- name: CreateTransaction :one
INSERT INTO transactions (group_id, created_by, kind)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetTransaction :one
SELECT * FROM transactions
WHERE id = $1;

-- name: GetTransactionsByGroup :many
SELECT * FROM transactions
WHERE group_id = $1
ORDER BY updated_at DESC;

-- name: UpdateTransaction :one
UPDATE transactions
SET updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteTransaction :exec
DELETE FROM transactions
WHERE id = $1;

-- name: CreatePayment :one
INSERT INTO payments (transaction_id, paid_by, paid_to, amount)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdatePayment :one
UPDATE payments
SET amount = $2, paid_by = $3, paid_to = $4
WHERE id = $1
RETURNING *;

-- name: GetPaymentsByGroup :many
SELECT * FROM payments
INNER JOIN transactions ON payments.transaction_id = transactions.id
WHERE transactions.group_id = $1
ORDER BY transactions.updated_at;
