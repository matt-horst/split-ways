package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
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
	}{}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Printf("Couldn't decode group data: %v\n", err)
		http.Error(w, "Couldn't create group", http.StatusBadRequest)
		return
	}

	group, err := cfg.Db.CreateGroup(
		r.Context(),
		database.CreateGroupParams{
			Name:  data.Name,
			Owner: user.ID,
		})
	if err != nil {
		log.Printf("Couldn't create group: %v\n", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	_, err = cfg.Db.CreateUserGroup(
		r.Context(),
		database.CreateUserGroupParams{
			UserID:  user.ID,
			GroupID: group.ID,
		},
	)
	if err != nil {
		log.Printf("Couldn't add user to group: %v\n", err)
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(group)
	if err != nil {
		log.Printf("Couldn't send response body: %v\n", err)
		return
	}
}

func (cfg *Config) HandlerAddUserToGroup(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(userContextKey).(database.User)
	if !ok {
		log.Printf("Attempted add user to group with unauthenticated user\n")
		http.Error(w, "User not authorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	groupIDPath, ok := vars["group_id"]
	if !ok {
		log.Printf("Missing group id\n")
		http.Error(w, "Missing group id", http.StatusBadRequest)
		return
	}

	groupID, err := uuid.Parse(groupIDPath)
	if err != nil {
		log.Printf("Couldn't parse group id: %v\n", err)
		http.Error(w, "Couldnt parse group id", http.StatusBadRequest)
		return
	}

	_, err = cfg.Db.CreateUserGroup(
		r.Context(),
		database.CreateUserGroupParams{
			UserID:  user.ID,
			GroupID: groupID,
		},
	)
	if err != nil {
		if strings.Contains(err.Error(), "users_groups_group_id_fkey") {
			log.Printf("Couldn't find group: %v\n", err)
			http.Error(w, "Couldn't find group", http.StatusBadRequest)
			return
		}

		log.Printf("Couldn't add user to group: %v\n", err)
		http.Error(w, "Couldn't add user to group", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
