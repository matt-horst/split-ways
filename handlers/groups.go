package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/matt-horst/split-ways/internal/database"
)

type exportUser struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

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

func (cfg *Config) HandlerUpdateGroup(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(userContextKey).(database.User)
	if !ok {
		log.Printf("Attempted to create group with unauthenticated user\n")
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

	group, err := cfg.Db.GetGroup(r.Context(), groupID)
	if err != nil {
		log.Printf("Couldn't find group: %v\n", err)
		http.Error(w, "Couldn't find group", http.StatusBadRequest)
		return
	}

	if group.Owner != user.ID {
		log.Printf("Attempt to update group from non-owner\n")
		http.Error(w, "Can't update group as non-owner", http.StatusForbidden)
		return
	}

	data := struct {
		Name string `json:"name"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		log.Printf("Couldn't decode group data: %v\n", err)
		http.Error(w, "Couldn't create group", http.StatusBadRequest)
		return
	}

	if _, err := cfg.Db.UpdateGroupName(
		r.Context(),
		database.UpdateGroupNameParams{ID: groupID, Name: data.Name},
	); err != nil {
		log.Printf("Couldn't update group: %v\n", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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

	_, err = cfg.Db.GetUserGroup(
		r.Context(),
		database.GetUserGroupParams{
			UserID:  user.ID,
			GroupID: groupID,
		},
	)
	if err != nil {
		log.Printf("Attempt to add user to group from unauthorized user: %v\n", err)
		http.Error(w, "User not member of group", http.StatusForbidden)
		return
	}

	data := struct {
		Username string `json:"username"`
	}{}

	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Printf("Couldn't decode request body: %v\n", err)
		http.Error(w, "Couldn't decode request", http.StatusBadRequest)
		return
	}

	addUser, err := cfg.Db.GetUserByUsername(r.Context(), data.Username)
	if err != nil {
		log.Printf("Couldn't find user: %v\n", err)
		http.Error(w, "Couldn't find user", http.StatusBadRequest)
		return
	}

	_, err = cfg.Db.CreateUserGroup(
		r.Context(),
		database.CreateUserGroupParams{
			UserID:  addUser.ID,
			GroupID: groupID,
		},
	)
	if err != nil {
		if strings.Contains(err.Error(), "users_groups_group_id_fkey") {
			log.Printf("Couldn't find group: %v\n", err)
			http.Error(w, "Couldn't find group", http.StatusBadRequest)
			return
		}

		// Duplicate key error
		if strings.Contains(err.Error(), "users_groups_user_id_group_id_key") {
			log.Printf("Couldn't add duplicate user group: %v\n", err)
			http.Error(w, "User already in group", http.StatusBadRequest)
			return
		}

		log.Printf("Couldn't add user to group: %v\n", err)
		http.Error(w, "Couldn't add user to group", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cfg *Config) HandlerGetGroupUsers(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(userContextKey).(database.User)
	if !ok {
		log.Printf("Attempted to get users for group by unauthorized user\n")
		http.Error(w, "User not authorized", http.StatusUnauthorized)
		return
	}

	groupIDPath, ok := mux.Vars(r)["group_id"]
	if !ok {
		log.Printf("Missing group id\n")
		http.Error(w, "Missing group id", http.StatusBadRequest)
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
		log.Printf("Attempt to read users for group user does not belong: %v\n", err)
		http.Error(w, "Not member of group", http.StatusForbidden)
		return
	}

	users, err := cfg.Db.GetUsersByGroup(
		r.Context(),
		groupID,
	)
	if err != nil {
		log.Printf("Couldn't find group: %v\n", err)
		http.Error(w, "Couldn't find group", http.StatusBadRequest)
		return
	}

	sanitizedUsers := make([]exportUser, len(users))

	for i, user := range users {
		sanitizedUsers[i] = exportUser{
			ID:        user.ID,
			Username:  user.Username,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(users)
	if err != nil {
		log.Printf("Couldn't send response body: %v\n", err)
		return
	}
}

func (cfg *Config) HandlerRemoveUserFromGroup(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(userContextKey).(database.User)
	if !ok {
		log.Printf("Attempt to remove user from group using unauthorized user\n")
		http.Error(w, "User unauthorized", http.StatusUnauthorized)
		return
	}

	groupIDPath, ok := mux.Vars(r)["group_id"]
	if !ok {
		log.Printf("Missing group id\n")
		http.Error(w, "Missing group id", http.StatusBadRequest)
		return
	}

	groupID, err := uuid.Parse(groupIDPath)
	if err != nil {
		log.Printf("Couldn't parse group id: %v\n", err)
		http.Error(w, "Couldn't parse group id", http.StatusBadRequest)
		return
	}

	group, err := cfg.Db.GetGroup(r.Context(), groupID)
	if err != nil {
		log.Printf("Couldn't find group: %v\n", err)
		http.Error(w, "Couldn't find group", http.StatusBadRequest)
		return
	}

	if group.Owner != user.ID {
		log.Printf("Attempt to remove user from group by non-owner\n")
		http.Error(w, "Can't remove users from group as non-owner", http.StatusForbidden)
		return
	}

	data := struct {
		ID uuid.UUID `json:"id"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		log.Printf("Couldn't decode request body: %v\n", err)
		http.Error(w, "Malformed request body", http.StatusBadRequest)
		return
	}

	if _, err := cfg.Db.DeleteUserGroup(r.Context(), database.DeleteUserGroupParams{UserID: data.ID, GroupID: groupID}); err != nil {
		log.Printf("Couln't remove user from group; user not in group: %v\n", err)
		http.Error(w, "User not in group", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
