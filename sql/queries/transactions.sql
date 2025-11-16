-- name: CreateExpense :one
INSERT INTO expenses (transaction_id, paid_by, description, amount)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateExpense :one
UPDATE expenses
SET paid_by = $2, description = $3
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

-- name: CreateTransaction :one
INSERT INTO transactions (group_id, created_by, kind)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetTransactionsByGroup :many
SELECT * FROM transactions
WHERE group_id = $1
ORDER BY updated_at;

-- name: CreatePayment :one
INSERT INTO payments (transaction_id, paid_by, paid_to, amount)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdatePayment :one
UPDATE payments
SET amount = $2
WHERE id = $1
RETURNING *;

-- name: GetPaymentsByGroup :many
SELECT * FROM payments
INNER JOIN transactions ON payments.transaction_id = transactions.id
WHERE transactions.group_id = $1
ORDER BY transactions.updated_at;
