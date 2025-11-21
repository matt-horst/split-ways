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

func (cfg *Config) HandlerCreatePayment(w http.ResponseWriter, r *http.Request) {
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

	_, err = cfg.Queries.GetUserGroup(
		r.Context(),
		database.GetUserGroupParams{
			UserID:  user.ID,
			GroupID: groupID,
		},
	)
	if err != nil {
		log.Printf("Attempt to create payment non-user group: %v\n", err)
		http.Error(w, "User does not belong to group", http.StatusForbidden)
		return
	}

	data := struct {
		PaidBy string          `json:"paid_by"`
		PaidTo string          `json:"paid_to"`
		Amount decimal.Decimal `json:"amount"`
	}{}

	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Printf("Couldn't decode request body: %v\n", err)
		http.Error(w, "Couldn't create payment", http.StatusBadRequest)
		return
	}

	paidBy, err := cfg.Queries.GetUserByUsername(r.Context(), data.PaidBy)
	if err != nil {
		log.Printf("Couldn't find paid by user: %v\n", err)
		http.Error(w, fmt.Sprintf("Couldn't find user `%v`", data.PaidBy), http.StatusBadRequest)
		return
	}

	paidTo, err := cfg.Queries.GetUserByUsername(r.Context(), data.PaidTo)
	if err != nil {
		log.Printf("Couldn't find paid by user: %v\n", err)
		http.Error(w, fmt.Sprintf("Couldn't find user `%v`", data.PaidTo), http.StatusBadRequest)
		return
	}

	_, err = cfg.Queries.GetUserGroup(
		r.Context(),
		database.GetUserGroupParams{
			UserID:  paidBy.ID,
			GroupID: groupID,
		},
	)
	if err != nil {
		log.Printf("Attempt to create payment paid by user not in group: %v\n", err)
		http.Error(w, fmt.Sprintf("User `%v` not in group", data.PaidBy), http.StatusBadRequest)
		return
	}

	_, err = cfg.Queries.GetUserGroup(
		r.Context(),
		database.GetUserGroupParams{
			UserID:  paidTo.ID,
			GroupID: groupID,
		},
	)
	if err != nil {
		log.Printf("Attempt to create payment paid to user not in group: %v\n", err)
		http.Error(w, fmt.Sprintf("User `%v` not in group", data.PaidTo), http.StatusBadRequest)
		return
	}

	transaction, err := cfg.Queries.CreateTransaction(
		r.Context(),
		database.CreateTransactionParams{
			GroupID: groupID,
			CreatedBy: uuid.NullUUID{
				UUID:  user.ID,
				Valid: true,
			},
			Kind: "payment",
		},
	)
	if err != nil {
		log.Printf("Couldn't create transaction: %v\n", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	payment, err := cfg.Queries.CreatePayment(
		r.Context(),
		database.CreatePaymentParams{
			TransactionID: transaction.ID,
			PaidBy: uuid.NullUUID{
				UUID:  paidBy.ID,
				Valid: true,
			},
			PaidTo: uuid.NullUUID{
				UUID:  paidTo.ID,
				Valid: true,
			},
			Amount: data.Amount,
		},
	)
	if err != nil {
		log.Printf("Couldn't create payment: %v\n", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(payment)
	if err != nil {
		log.Printf("Couldn't send response body: %v\n", err)
		return
	}
}

func (cfg *Config) HandlerUpdatePayment(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(userContextKey).(database.User)
	if !ok {
		log.Printf("Attempted to update payment by unauthorized user\n")
		http.Error(w, "User not authorized", http.StatusUnauthorized)
		return
	}

	txID, err := uuid.Parse(r.URL.Query().Get("id"))
	if err != nil {
		log.Printf("Couldn't parse transaction ID: %v\n", err)
		http.Error(w, "Couldn't find transaction", http.StatusBadRequest)
		return
	}

	tx, err := cfg.Queries.GetTransaction(r.Context(), txID)
	if err != nil {
		log.Printf("Couldn't find transaction: %v\n", err)
		http.Error(w, "Couldn't find transaction", http.StatusBadRequest)
		return
	}

	if !tx.CreatedBy.Valid || tx.CreatedBy.UUID != user.ID {
		log.Printf("Attempt to update payment for other user\n")
		http.Error(w, "Can't update other users' payemnts", http.StatusForbidden)
		return
	}

	payment, err := cfg.Queries.GetPaymentByTransaction(r.Context(), txID)
	if err != nil {
		log.Printf("Couldn't find payment: %v\n", err)
		http.Error(w, "Couldn't find payment", http.StatusBadRequest)
		return
	}

	data := struct {
		PaidBy string          `json:"paid_by"`
		PaidTo string          `json:"paid_to"`
		Amount decimal.Decimal `json:"amount"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		log.Printf("Couldn't decode request body: %v\n", err)
		http.Error(w, "Couldn't update payment", http.StatusBadRequest)
		return
	}

	paidBy := payment.PaidBy
	if data.PaidBy != "" {
		paidByUser, err := cfg.Queries.GetUserByUsername(r.Context(), data.PaidBy)
		if err != nil {
			log.Printf("Couln't find user: %v\n", err)
			http.Error(w, fmt.Sprintf("Couldn't find user %s", data.PaidBy), http.StatusBadRequest)
			return
		}

		paidBy.UUID = paidByUser.ID
		paidBy.Valid = true
	}

	paidTo := payment.PaidTo
	if data.PaidTo != "" {
		paidToUser, err := cfg.Queries.GetUserByUsername(r.Context(), data.PaidTo)
		if err != nil {
			log.Printf("Couln't find user: %v\n", err)
			http.Error(w, fmt.Sprintf("Couldn't find user %s", data.PaidTo), http.StatusBadRequest)
			return
		}

		paidTo.UUID = paidToUser.ID
		paidTo.Valid = true
	}

	if data.Amount.LessThan(decimal.Zero) {
		data.Amount = payment.Amount
	}

	payment, err = cfg.Queries.UpdatePayment(
		r.Context(),
		database.UpdatePaymentParams{
			ID:     payment.ID,
			Amount: data.Amount,
			PaidBy: paidBy,
			PaidTo: paidTo,
		},
	)
	if err != nil {
		log.Printf("Couldn't update payment: %v\n", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
