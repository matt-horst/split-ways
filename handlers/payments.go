package handlers

import (
	"encoding/json"
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
		PaidBy uuid.UUID `json:"paid_by"`
		PaidTo uuid.UUID `json:"paid_to"`
		Amount string    `json:"amount"`
	}{}

	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Printf("Couldn't decode request body: %v\n", err)
		http.Error(w, "Couldn't create payment", http.StatusBadRequest)
		return
	}

	_, err = cfg.Db.GetUserGroup(
		r.Context(),
		database.GetUserGroupParams{
			UserID:  data.PaidBy,
			GroupID: groupID,
		},
	)
	if err != nil {
		log.Printf("Attempt to create payment paid by user not in group: %v\n", err)
		http.Error(w, "Couldn't create payment", http.StatusBadRequest)
		return
	}

	_, err = cfg.Db.GetUserGroup(
		r.Context(),
		database.GetUserGroupParams{
			UserID:  data.PaidTo,
			GroupID: groupID,
		},
	)
	if err != nil {
		log.Printf("Attempt to create payment paid to user not in group: %v\n", err)
		http.Error(w, "Couldn't create payment", http.StatusBadRequest)
		return
	}

	payment, err := cfg.Db.CreatePayment(
		r.Context(),
		database.CreatePaymentParams{
			GroupID: groupID,
			CreatedBy: uuid.NullUUID{
				UUID:  user.ID,
				Valid: true,
			},
			PaidBy: uuid.NullUUID{
				UUID:  data.PaidBy,
				Valid: true,
			},
			PaidTo: uuid.NullUUID{
				UUID:  data.PaidTo,
				Valid: true,
			},
			Amount: data.Amount,
		},
	)
	if err != nil {
		log.Printf("Couldn't create expense: %v\n", err)
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
