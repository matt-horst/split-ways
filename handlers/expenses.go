package handlers

import (
	"encoding/json"
	"log"
	"net/http"

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
		Splits      []struct {
			UserID uuid.UUID `json:"user_id"`
			Amount string    `json:"amount"`
		} `json:"splits"`
	}{}

	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Printf("Couldn't decode request body: %v\n", err)
		http.Error(w, "Couldn't create expense", http.StatusBadRequest)
		return
	}

	expense, err := cfg.Db.CreateExpense(
		r.Context(),
		database.CreateExpenseParams{
			GroupID: groupID,
			CreatedBy: uuid.NullUUID{
				UUID:  user.ID,
				Valid: true,
			},
			PaidBy: uuid.NullUUID{
				UUID:  user.ID,
				Valid: true,
			},
			Description: data.Description,
		},
	)
	if err != nil {
		log.Printf("Couldn't create expense: %v\n", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	for _, split := range data.Splits {
		_, err = cfg.Db.CreateSplit(
			r.Context(),
			database.CreateSplitParams{
				ExpenseID: expense.ID,
				UserID: uuid.NullUUID{
					UUID:  split.UserID,
					Valid: true,
				},
				Amount: split.Amount,
			},
		)
		if err != nil {
			log.Printf("Couldn't create split: %v\n", err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)

			_, err = cfg.Db.DeleteExpense(r.Context(), expense.ID)
			if err != nil {
				log.Printf("Couldn't clean up invalid expense: %v\n", err)
				return
			}

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
