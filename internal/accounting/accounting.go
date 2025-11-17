package accounting

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/matt-horst/split-ways/internal/database"
	"github.com/shopspring/decimal"
)

type TransactionKind string

const (
	ExpenseKind TransactionKind = "expense"
	PaymentKind TransactionKind = "payment"
)

type User struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
}

type Transaction struct {
	ID        uuid.UUID       `json:"id"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	CreatedBy *User           `json:"created_by"`
	Kind      TransactionKind `json:"kind"`
	Payment   *Payment        `json:"payment"`
	Expense   *Expense        `json:"expense"`
}

type Payment struct {
	PaidBy *User           `json:"paid_by"`
	PaidTo *User           `json:"paid_to"`
	Amount decimal.Decimal `json:"amount"`
}

type Expense struct {
	Description string          `json:"description"`
	PaidBy      *User           `json:"paid_by"`
	Amount      decimal.Decimal `json:"amount"`
	Debts       []Debt          `json:"debts"`
}

type Debt struct {
	Amount decimal.Decimal `json:"amount"`
	OwedBy *User           `json:"owed_by"`
	OwedTo *User           `json:"owed_to"`
}

func GetTransationsByGroup(queries *database.Queries, ctx context.Context, groupID uuid.UUID) ([]Transaction, error) {
	dbTransactions, err := queries.GetTransactionsByGroup(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("couldn't get transactions by group: %v", err)
	}

	users := make(map[uuid.UUID]*User)

	dbUsers, err := queries.GetUsersByGroup(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("coudln't find group: %v", err)
	}

	for _, u := range dbUsers {
		users[u.ID] = &User{ID: u.ID, Username: u.Username}
	}

	transactions := make([]Transaction, len(dbTransactions))

	for i, dbTransaction := range dbTransactions {
		var createdByUser *User
		if dbTransaction.CreatedBy.Valid {
			if u, ok := users[dbTransaction.CreatedBy.UUID]; ok {
				createdByUser = u
			}
		}

		switch dbTransaction.Kind {
		case "expense":
			dbExpense, err := queries.GetExpenseByTransaction(ctx, dbTransaction.ID)
			if err != nil {
				return nil, fmt.Errorf("couldn't find payment for transaction: %v", err)
			}

			dbDebts, err := queries.GetDebtsByTransaction(ctx, dbTransaction.ID)
			if err != nil {
				return nil, fmt.Errorf("couldn't get debts for transaction: %v", err)
			}

			var paidByUser *User
			if dbExpense.PaidBy.Valid {
				if u, ok := users[dbExpense.PaidBy.UUID]; ok {
					paidByUser = u
				}
			}

			expense := &Expense{
				Description: dbExpense.Description,
				PaidBy:      paidByUser,
				Amount:      dbExpense.Amount,
				Debts:       make([]Debt, len(dbDebts)),
			}

			for j, dbDebt := range dbDebts {
				var owedByUser *User
				if dbDebt.OwedBy.Valid {
					if u, ok := users[dbDebt.OwedBy.UUID]; ok {
						owedByUser = u
					}
				}

				var owedToUser *User
				if dbDebt.OwedTo.Valid {
					if u, ok := users[dbDebt.OwedTo.UUID]; ok {
						owedToUser = u
					}
				}

				debt := Debt{
					Amount: dbDebt.Amount,
					OwedBy: owedByUser,
					OwedTo: owedToUser,
				}

				expense.Debts[j] = debt
			}

			transactions[i] = Transaction{
				ID:        dbTransaction.ID,
				CreatedAt: dbTransaction.CreatedAt,
				UpdatedAt: dbTransaction.UpdatedAt,
				CreatedBy: createdByUser,
				Kind:      TransactionKind(dbTransaction.Kind),
				Expense:   expense,
			}
		case "payment":
			dbPayment, err := queries.GetPaymentByTransaction(ctx, dbTransaction.ID)
			if err != nil {
				return nil, fmt.Errorf("couldn't find payment for transaction: %v", err)
			}

			var paidByUser *User
			if dbPayment.PaidBy.Valid {
				if u, ok := users[dbPayment.PaidBy.UUID]; ok {
					paidByUser = u
				}
			}

			var paidToUser *User
			if dbPayment.PaidTo.Valid {
				if u, ok := users[dbPayment.PaidTo.UUID]; ok {
					paidToUser = u
				}
			}

			payment := &Payment{
				PaidBy: paidByUser,
				PaidTo: paidToUser,
				Amount: dbPayment.Amount,
			}

			transactions[i] = Transaction{
				ID:        dbTransaction.ID,
				CreatedAt: dbTransaction.CreatedAt,
				UpdatedAt: dbTransaction.UpdatedAt,
				CreatedBy: createdByUser,
				Kind:      TransactionKind(dbTransaction.Kind),
				Payment:   payment,
			}
		default:
			return nil, fmt.Errorf("unknown transaction kind: %v", dbTransaction.Kind)
		}
	}

	return transactions, nil
}

type Balance struct {
	Other  User
	Amount decimal.Decimal
}

func GetBalanceForGroup(queries *database.Queries, ctx context.Context, groupID, userId uuid.UUID) ([]Balance, error) {
	users, err := queries.GetUsersByGroup(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("coudln't find users in group: %v", err)
	}

	balances := make([]Balance, len(users)-1)
	i := 0
	for _, user := range users {
		if user.ID == userId {
			continue
		}

		b, err := GetBalanceBetweenUsers(queries, ctx, groupID, userId, user.ID)
		if err != nil {
			return nil, err
		}

		balances[i] = Balance{
			Other:  User{ID: user.ID, Username: user.Username},
			Amount: b,
		}

		i++
	}

	return balances, nil
}

func GetBalanceBetweenUsers(queries *database.Queries, ctx context.Context, groupID, thisUserID, otherUserID uuid.UUID) (decimal.Decimal, error) {
	thisUserNullID := uuid.NullUUID{
		UUID:  thisUserID,
		Valid: true,
	}
	otherUserNullID := uuid.NullUUID{
		UUID:  otherUserID,
		Valid: true,
	}

	totalDebtToOther, err := queries.GetSumOfDebts(
		ctx,
		database.GetSumOfDebtsParams{
			GroupID: groupID,
			OwedBy:  thisUserNullID,
			OwedTo:  otherUserNullID,
		},
	)
	if err != nil {
		if !strings.Contains(err.Error(), "NULL") {
			return decimal.Decimal{}, fmt.Errorf("couldn't get debts between users: %v", err)
		}
	}

	totalDebtToThis, err := queries.GetSumOfDebts(
		ctx,
		database.GetSumOfDebtsParams{
			GroupID: groupID,
			OwedBy:  otherUserNullID,
			OwedTo:  thisUserNullID,
		},
	)
	if err != nil {
		if !strings.Contains(err.Error(), "NULL") {
			return decimal.Decimal{}, fmt.Errorf("couldn't get debts between users: %v", err)
		}
	}

	totalPaymentsToOther, err := queries.GetSumOfPayments(
		ctx,
		database.GetSumOfPaymentsParams{
			GroupID: groupID,
			PaidBy:  thisUserNullID,
			PaidTo:  otherUserNullID,
		},
	)
	if err != nil {
		if !strings.Contains(err.Error(), "NULL") {
			return decimal.Decimal{}, fmt.Errorf("couldn't get payments between users: %v", err)
		}
	}

	totalPaymentsToThis, err := queries.GetSumOfPayments(
		ctx,
		database.GetSumOfPaymentsParams{
			GroupID: groupID,
			PaidBy:  otherUserNullID,
			PaidTo:  thisUserNullID,
		},
	)
	if err != nil {
		if !strings.Contains(err.Error(), "NULL") {
			return decimal.Decimal{}, fmt.Errorf("couldn't get payments between users: %v", err)
		}
	}

	total := totalDebtToThis.Sub(totalDebtToOther).Add(totalPaymentsToOther).Sub(totalPaymentsToThis)

	return total, nil
}
