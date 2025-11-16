package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/matt-horst/split-ways/internal/database"
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

	_, err = cfg.Db.GetUserGroup(
		r.Context(),
		database.GetUserGroupParams{
			UserID:  user.ID,
			GroupID: groupID,
		},
	)
	if err != nil {
		log.Printf("Attempt to create expense non-user group: %v\n", err)
		http.Error(w, "User does not belong to group", http.StatusForbidden)
		return
	}

	data := struct {
		Description string `json:"description"`
		Amount      string `json:"amount"`
	}{}

	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Printf("Couldn't decode request body: %v\n", err)
		http.Error(w, "Couldn't create expense", http.StatusBadRequest)
		return
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
			PaidBy: uuid.NullUUID{
				UUID:  user.ID,
				Valid: true,
			},
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

	amount, err := strconv.Atoi(strings.ReplaceAll(data.Amount, ".", ""))
	if err != nil {
		log.Printf("Couldn't convert amount to int: %v", err)
		http.Error(w, "Couldn't parse amount", http.StatusBadRequest)
		return
	}
	debtAmount := amount / (len(users) - 1)

	for _, u := range users {
		if u.ID == user.ID {
			continue
		}

		_, err = cfg.Db.CreateDebt(
			r.Context(),
			database.CreateDebtParams{
				ExpenseID: expense.ID,
				OwedTo: uuid.NullUUID{
					UUID:  user.ID,
					Valid: true,
				},
				OwedBy: uuid.NullUUID{
					UUID:  u.ID,
					Valid: true,
				},
				Amount: fmt.Sprintf("%d.%d", debtAmount/100, debtAmount%100),
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
