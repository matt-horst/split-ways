package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/matt-horst/split-ways/internal/database"
)

func (cfg *Config) HandlerDeleteTransaction(w http.ResponseWriter, r *http.Request) {
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
		log.Printf("Attempt to create expense non-user group: %v\n", err)
		http.Error(w, "User does not belong to group", http.StatusForbidden)
		return
	}

	data := struct {
		ID uuid.UUID `json:"id"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		log.Printf("Couldn't decode request body: %v\n", err)
		http.Error(w, "Couldn't decode request body", http.StatusBadRequest)
		return
	}

	tx, err := cfg.Queries.GetTransaction(r.Context(), data.ID)
	if err != nil {
		log.Printf("Couldn't find transaction: %v\n", err)
		http.Error(w, "Couldn't find transaction", http.StatusBadRequest)
		return
	}

	if !tx.CreatedBy.Valid || tx.CreatedBy.UUID != user.ID {
		log.Printf("Attempt to delete transaction by unauthroized user: %v\n", err)
		http.Error(w, "You do not own this transaction", http.StatusForbidden)
		return
	}

	if err := cfg.Queries.DeleteTransaction(r.Context(), data.ID); err != nil {
		log.Printf("Couldn't delete transaction: %v\n", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
