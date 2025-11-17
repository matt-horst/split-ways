-- name: GetUserBalanceByGroup :one
SELECT SUM(payments.amount) - SUM(debts.amount) AS balance FROM transactions
INNER JOIN expenses ON transactions.id = expenses.transaction_id
INNER JOIN debts ON expenses.id = debts.expense_id
INNER JOIN payments ON transactions.id = payments.transaction_id
WHERE transactions.group_id = $1
AND debts.owed_to = $3
AND debts.owed_by = $2
AND payments.paid_to = $3
AND payments.paid_by = $2;

-- name: GetSumOfDebts :one
SELECT CAST(COALESCE(SUM(debts.amount), 0) AS NUMERIC(12, 2)) AS total FROM transactions
INNER JOIN expenses ON transactions.id = expenses.transaction_id
INNER JOIN debts ON expenses.id = debts.expense_id
WHERE transactions.group_id = $1 AND debts.owed_by = $2 AND debts.owed_to = $3;

-- name: GetSumOfPayments :one
SELECT CAST(COALESCE(SUM(payments.amount), 0) AS NUMERIC(12, 2)) AS total FROM transactions
INNER JOIN payments ON transactions.id = payments.transaction_id
WHERE transactions.group_id = $1 AND payments.paid_by = $2 AND payments.paid_to = $3;

-- name: GetExpenseByTransaction :one
SELECT * FROM expenses
WHERE expenses.transaction_id = $1;

-- name: GetDebtsByTransaction :many
SELECT debts.* FROM expenses
INNER JOIN debts ON expenses.id = debts.expense_id
WHERE expenses.transaction_id = $1;

-- name: GetPaymentByTransaction :one
SELECT payments.* FROM payments
WHERE payments.transaction_id = $1;
