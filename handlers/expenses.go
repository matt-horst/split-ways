package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/matt-horst/split-ways/internal/database"
	"github.com/shopspring/decimal"
)

func (cfg *Config) HandlerCreateExpense(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(userContextKey).(database.User)
	if !ok {
		log.Printf("Attempted to create expense with unauthenticated user")
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	groupIDPath, ok := mux.Vars(r)["group_id"]
	if !ok {
		log.Printf("Couldn't find group id\n")
		http.Error(w, "Couldn't find group id", http.StatusBadRequest)
		return
	}

	groupID, err := uuid.Parse(groupIDPath)
	if err != nil {
		log.Printf("Couldn't parse group id: %v\n", err)
		http.Error(w, "Couldn't parse group id", http.StatusBadRequest)
		return
	}

	if _, err = cfg.Db.GetUserGroup(
		r.Context(),
		database.GetUserGroupParams{
			UserID:  user.ID,
			GroupID: groupID,
		},
	); err != nil {
		log.Printf("Attempt to create expense non-user group: %v\n", err)
		http.Error(w, "User does not belong to group", http.StatusForbidden)
		return
	}

	data := struct {
		Description string          `json:"description"`
		Amount      decimal.Decimal `json:"amount"`
		PaidBy      string          `json:"paid_by"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		log.Printf("Couldn't decode request body: %v\n", err)
		http.Error(w, "Couldn't create expense", http.StatusBadRequest)
		return
	}

	paidBy := uuid.NullUUID{Valid: true, UUID: user.ID}
	if data.PaidBy != "" {
		paidByUser, err := cfg.Db.GetUserByUsername(r.Context(), data.PaidBy)
		if err != nil {
			log.Printf("Couldn't find user: %v\n", err)
			http.Error(w, fmt.Sprintf("Couldn't find user %s", data.PaidBy), http.StatusBadRequest)
			return
		}

		if _, err := cfg.Db.GetUserGroup(
			r.Context(),
			database.GetUserGroupParams{GroupID: groupID, UserID: paidByUser.ID},
		); err != nil {
			log.Printf("Couldn't create expense where paid by user is not in group: %v\n", err)
			http.Error(w, fmt.Sprintf("%s is not in group", data.PaidBy), http.StatusBadRequest)
			return
		}

		paidBy.UUID = paidByUser.ID
	}

	transaction, err := cfg.Db.CreateTransaction(
		r.Context(),
		database.CreateTransactionParams{
			GroupID: groupID,
			CreatedBy: uuid.NullUUID{
				UUID:  user.ID,
				Valid: true,
			},
			Kind: "expense",
		},
	)
	if err != nil {
		log.Printf("Couldn't create transaction: %v\n", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	expense, err := cfg.Db.CreateExpense(
		r.Context(),
		database.CreateExpenseParams{
			TransactionID: transaction.ID,
			PaidBy: paidBy,
			Description: data.Description,
			Amount:      data.Amount,
		},
	)
	if err != nil {
		log.Printf("Couldn't create expense: %v\n", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	users, err := cfg.Db.GetUsersByGroup(r.Context(), groupID)
	if err != nil {
		log.Printf("Couldn't get users in group: %v", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	debtAmount := data.Amount.Div(decimal.NewFromInt(int64(len(users))))

	for _, u := range users {
		if u.ID == expense.PaidBy.UUID {
			continue
		}

		_, err = cfg.Db.CreateDebt(
			r.Context(),
			database.CreateDebtParams{
				ExpenseID: expense.ID,
				OwedTo: expense.PaidBy,
				OwedBy: uuid.NullUUID{
					UUID:  u.ID,
					Valid: true,
				},
				Amount: debtAmount,
			},
		)
		if err != nil {
			log.Printf("Couldn't create split: %v\n", err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(expense)
	if err != nil {
		log.Printf("Couldn't send response body: %v\n", err)
		return
	}
}

func (cfg *Config) HandlerEditExpense(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(userContextKey).(database.User)
	if !ok {
		log.Printf("Attempt to edit expense for unauthorized user\n")
		http.Error(w, "User unauthorized", http.StatusUnauthorized)
		return
	}

	id, err := uuid.Parse(r.URL.Query().Get("id"))
	if err != nil {
		log.Printf("Couldn't parse expense id: %v\n", err)
		http.Error(w, "Couldn't find expense", http.StatusBadRequest)
		return
	}

	tx, err := cfg.Db.GetTransaction(r.Context(), id)
	if err != nil {
		log.Printf("Couldn't get transaction: %v\n", err)
		http.Error(w, "Couldn't find transaction", http.StatusBadRequest)
		return
	}

	if !tx.CreatedBy.Valid || tx.CreatedBy.UUID != user.ID {
		log.Printf("Attempt to edit expense for unauthorized user\n")
		http.Error(w, "Can't edit other users' expenses", http.StatusForbidden)
		return
	}

	expense, err := cfg.Db.GetExpenseByTransaction(r.Context(), tx.ID)
	if err != nil {
		log.Printf("Couldn't find expense by transaction: %v\n", err)
		http.Error(w, "Couldn't find expense", http.StatusBadRequest)
		return
	}

	data := struct {
		Amount      decimal.Decimal `json:"amount"`
		Description string          `json:"description"`
		PaidBy      string          `json:"paid_by"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		log.Printf("Couldn't decode request body: %v\n", err)
		http.Error(w, "Malformed request", http.StatusBadRequest)
		return
	}

	if data.Amount.LessThan(decimal.Zero) {
		data.Amount = expense.Amount
	}

	if data.Description == "" {
		data.Description = expense.Description
	}

	paidByID := expense.PaidBy
	if data.PaidBy != "" {
		paidBy, err := cfg.Db.GetUserByUsername(r.Context(), data.PaidBy)
		if err != nil {
			log.Printf("Couldn't find user: %v\n", err)
			http.Error(w, fmt.Sprintf("Couldn't find user %s", data.PaidBy), http.StatusBadRequest)
			return
		}

		paidByID.Valid = true
		paidByID.UUID = paidBy.ID
	}

	expense, err = cfg.Db.UpdateExpense(
		r.Context(),
		database.UpdateExpenseParams{
			ID:          expense.ID,
			PaidBy:      paidByID,
			Description: data.Description,
			Amount:      data.Amount,
		},
	)
	if err != nil {
		log.Printf("Couldn't update expense: %v\n", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	usersInGroup, err := cfg.Db.GetUsersByGroup(r.Context(), tx.GroupID)
	if err != nil {
		log.Printf("Couldn't get users in group: %v\n", err)
		http.Error(w, "Couldn't find group", http.StatusBadRequest)
		return
	}

	debtAmount := expense.Amount.Div(decimal.NewFromInt(int64(len(usersInGroup))))

	if err := cfg.Db.DeleteDebtsByExpense(r.Context(), expense.ID); err != nil {
		log.Printf("Couldn't delete debts for expense: %v\n", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	for _, u := range usersInGroup {
		if expense.PaidBy.Valid && expense.PaidBy.UUID == u.ID {
			continue
		}

		if _, err := cfg.Db.CreateDebt(
			r.Context(),
			database.CreateDebtParams{
				ExpenseID: expense.ID,
				OwedTo:    expense.PaidBy,
				OwedBy:    uuid.NullUUID{Valid: true, UUID: u.ID},
				Amount:    debtAmount,
			},
		); err != nil {
			log.Printf("Couldn't create debt: %v\n", err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}
	}

	if _, err := cfg.Db.UpdateTransaction(r.Context(), tx.ID); err != nil {
		log.Printf("Couldn't update transaction: %v\n", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
