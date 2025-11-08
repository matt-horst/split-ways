-- name: GetUserBalanceByGroup :one
SELECT users_groups.group_id AS group_id, SUM(payments.amount) - SUM(splits.amount) AS balance FROM splits
INNER JOIN expenses ON splits.expense_id = expenses.id
INNER JOIN users_groups ON expenses.group_id = users_groups.group_id
INNER JOIN payments ON payments.group_id = users_groups.group_id AND payments.paid_to = $1
WHERE users_groups.user_id = $1
GROUP BY users_groups.group_id;
