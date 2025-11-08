package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/matt-horst/split-ways/internal/database"
)

func (cfg *Config) HandlerCreateGroup(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(userContextKey).(database.User)
	if !ok {
		log.Printf("Attempted to create group with unauthenticated user\n")
		http.Error(w, "User not authorized", http.StatusUnauthorized)
		return
	}

	data := struct {
		Name string `json:"name"`
	} {}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Printf("Couldn't decode group data: %v\n", err)
		http.Error(w, "Couldn't create group", http.StatusBadRequest)
		return
	}

	group, err := cfg.Db.CreateGroup(
		r.Context(),
		database.CreateGroupParams{
			Name: data.Name,
			Owner: user.ID,
	})
	if err != nil {
		log.Printf("Couldn't create group: %v\n", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(group)
	if err != nil {
		log.Printf("Couldn't send response body: %v\n", err)
		return
	}
}
