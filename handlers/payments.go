package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/matt-horst/split-ways/internal/database"
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

	_, err = cfg.Db.GetUserGroup(
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
		PaidBy string `json:"paid_by"`
		PaidTo string `json:"paid_to"`
		Amount string `json:"amount"`
	}{}

	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Printf("Couldn't decode request body: %v\n", err)
		http.Error(w, "Couldn't create payment", http.StatusBadRequest)
		return
	}

	paidBy, err := cfg.Db.GetUserByUsername(r.Context(), data.PaidBy)
	if err != nil {
		log.Printf("Couldn't find paid by user: %v\n", err)
		http.Error(w, fmt.Sprintf("Couldn't find user `%v`", data.PaidBy), http.StatusBadRequest)
		return
	}

	paidTo, err := cfg.Db.GetUserByUsername(r.Context(), data.PaidTo)
	if err != nil {
		log.Printf("Couldn't find paid by user: %v\n", err)
		http.Error(w, fmt.Sprintf("Couldn't find user `%v`", data.PaidTo), http.StatusBadRequest)
		return
	}

	_, err = cfg.Db.GetUserGroup(
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

	_, err = cfg.Db.GetUserGroup(
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

	transaction, err := cfg.Db.CreateTransaction(
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

	payment, err := cfg.Db.CreatePayment(
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
