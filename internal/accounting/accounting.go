package accounting

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/matt-horst/split-ways/internal/database"
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
	PaidBy *User  `json:"paid_by"`
	PaidTo *User  `json:"paid_to"`
	Amount string `json:"amount"`
}

type Expense struct {
	Description string `json:"description"`
	PaidBy      *User  `json:"paid_by"`
	Amount      string `json:"amount"`
	Debts       []Debt `json:"debts"`
}

type Debt struct {
	Amount string `json:"amount"`
	OwedBy *User  `json:"owed_by"`
	OwedTo *User  `json:"owed_to"`
}

func GetTransationsByGroup(queries *database.Queries, ctx context.Context, groupID uuid.UUID) ([]Transaction, error) {
	dbTransactions, err := queries.GetTransactionsByGroup(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("Couldn't get transactions by group: %v\n", err)
	}

	users := make(map[uuid.UUID]*User)

	dbUsers, err := queries.GetUsersByGroup(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("Coudln't find group: %v", err)
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
				return nil, fmt.Errorf("Couldn't find payment for transaction: %v", err)
			}

			dbDebts, err := queries.GetDebtsByTransaction(ctx, dbTransaction.ID)

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
				return nil, fmt.Errorf("Couldn't find payment for transaction: %v", err)
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
			return nil, fmt.Errorf("Unknown transaction kind: %v", dbTransaction.Kind)
		}
	}

	return transactions, nil
}

func GetBalanceBetweenUsers(queries *database.Queries, ctx context.Context, groupID, thisUserID, otherUserID uuid.UUID) (int64, error) {
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
		return 0, fmt.Errorf("Couldn't get debts between users: %v\n", err)
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
		return 0, fmt.Errorf("Couldn't get debts between users: %v\n", err)
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
		return 0, fmt.Errorf("Couldn't get payments between users: %v\n", err)
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
		return 0, fmt.Errorf("Couldn't get payments between users: %v\n", err)
	}

	total := totalDebtToThis - totalDebtToOther + totalPaymentsToOther - totalPaymentsToThis

	return total, nil
}
