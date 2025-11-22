package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/matt-horst/split-ways/internal/api"
	"github.com/matt-horst/split-ways/internal/database"
)

type ExportUser struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateGroupData struct {
	Name string `json:"name"`
}

type UpdateGroupData struct {
	Name string `json:"name"`
}

type AddUserToGroupData struct {
	Username string `json:"username"`
}

type RemoveUserFromGroupData struct {
	ID uuid.UUID `json:"id"`
}

func (cfg *Config) HandlerCreateGroup(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(userContextKey).(database.User)
	if !ok {
		log.Printf("Attempted to create group with unauthenticated user\n")
		http.Error(w, "User not authorized", http.StatusUnauthorized)
		return
	}

	data := CreateGroupData{}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Printf("Couldn't decode group data: %v\n", err)
		http.Error(w, "Couldn't create group", http.StatusBadRequest)
		return
	}

	group, err := api.CreateGroup(
		r.Context(),
		cfg.DB,
		cfg.Tx,
		cfg.Queries,
		database.CreateGroupParams{
			Name:  data.Name,
			Owner: user.ID,
		},
	)
	if err != nil {
		log.Printf("couldn't create group: %v\n", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(group)
	if err != nil {
		log.Printf("couldn't write response body: %v\n", err)
		return
	}
}

func (cfg *Config) HandlerUpdateGroup(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(userContextKey).(database.User)
	if !ok {
		log.Printf("attempted to create group with unauthenticated user\n")
		http.Error(w, "User not authorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	groupIDPath, ok := vars["group_id"]
	if !ok {
		log.Printf("update group request missing group id\n")
		http.Error(w, "Missing group id", http.StatusBadRequest)
		return
	}

	groupID, err := uuid.Parse(groupIDPath)
	if err != nil {
		log.Printf("couldn't parse group id: %v\n", err)
		http.Error(w, "Couldnt parse group id", http.StatusBadRequest)
		return
	}

	group, err := cfg.Queries.GetGroup(r.Context(), groupID)
	if err != nil {
		log.Printf("couldn't find group: %v\n", err)
		http.Error(w, "Couldn't find group", http.StatusBadRequest)
		return
	}

	if group.Owner != user.ID {
		log.Printf("attempt to update group from non-owner\n")
		http.Error(w, "Can't update group as non-owner", http.StatusForbidden)
		return
	}

	data := UpdateGroupData{}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		log.Printf("couldn't decode request body: %v\n", err)
		http.Error(w, "Couldn't create group", http.StatusBadRequest)
		return
	}

	group, err = cfg.Queries.UpdateGroupName(
		r.Context(),
		database.UpdateGroupNameParams{ID: groupID, Name: data.Name},
	)
	if err != nil {
		log.Printf("couldn't update group: %v\n", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(group); err != nil {
		log.Printf("couldn't write response body: %v\n", err)
	}
}

func (cfg *Config) HandlerAddUserToGroup(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(userContextKey).(database.User)
	if !ok {
		log.Printf("attempted add user to group with unauthenticated user\n")
		http.Error(w, "User not authorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	groupIDPath, ok := vars["group_id"]
	if !ok {
		log.Printf("add user to group request missing group id\n")
		http.Error(w, "Missing group id", http.StatusBadRequest)
		return
	}

	groupID, err := uuid.Parse(groupIDPath)
	if err != nil {
		log.Printf("couldn't parse group id: %v\n", err)
		http.Error(w, "Couldnt parse group id", http.StatusBadRequest)
		return
	}

	group, err := cfg.Queries.GetGroup(r.Context(), groupID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("couldn't find group: %v\n", err)
			http.Error(w, "Couldn't find group", http.StatusBadRequest)
			return
		}

		log.Printf("couldn't get group: %v\n", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	if group.Owner != user.ID {
		http.Error(w, "You must be group owner to perform this action", http.StatusForbidden)
		return
	}

	data := AddUserToGroupData{}

	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Printf("couldn't decode request body: %v\n", err)
		http.Error(w, "Couldn't decode request", http.StatusBadRequest)
		return
	}

	addUser, err := cfg.Queries.GetUserByUsername(r.Context(), data.Username)
	if err != nil {
		log.Printf("couldn't find user: %v\n", err)
		http.Error(w, "Couldn't find user", http.StatusBadRequest)
		return
	}

	_, err = cfg.Queries.CreateUserGroup(
		r.Context(),
		database.CreateUserGroupParams{
			UserID:  addUser.ID,
			GroupID: groupID,
		},
	)
	if err != nil {
		if strings.Contains(err.Error(), "users_groups_group_id_fkey") {
			log.Printf("couldn't find group: %v\n", err)
			http.Error(w, "Couldn't find group", http.StatusBadRequest)
			return
		}

		// Duplicate key error
		if strings.Contains(err.Error(), "users_groups_user_id_group_id_key") {
			log.Printf("couldn't add duplicate user group: %v\n", err)
			http.Error(w, "User already in group", http.StatusBadRequest)
			return
		}

		log.Printf("couldn't add user to group: %v\n", err)
		http.Error(w, "Soemthing went wrong", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cfg *Config) HandlerGetGroupUsers(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(userContextKey).(database.User)
	if !ok {
		log.Printf("attempted to get users for group by unauthorized user\n")
		http.Error(w, "User not authorized", http.StatusUnauthorized)
		return
	}

	groupIDPath, ok := mux.Vars(r)["group_id"]
	if !ok {
		log.Printf("get group users request missing group id\n")
		http.Error(w, "Missing group id", http.StatusBadRequest)
		return
	}

	groupID, err := uuid.Parse(groupIDPath)
	if err != nil {
		log.Printf("couldn't parse group id: %v\n", err)
		http.Error(w, "Couldn't parse group id", http.StatusBadRequest)
		return
	}

	ok, err = api.IsUserInGroup(r.Context(), cfg.Queries, groupID, user.ID)
	if err != nil {
		log.Printf("couldn't determine if user in group: %v\n", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(w, "User not member of group", http.StatusForbidden)
		return
	}

	users, err := cfg.Queries.GetUsersByGroup(
		r.Context(),
		groupID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("couldn't find group: %v\n", err)
			http.Error(w, "Couldn't find group", http.StatusBadRequest)
			return
		}

		log.Printf("couldn't get users by group: %v\n", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	sanitizedUsers := make([]ExportUser, len(users))

	for i, user := range users {
		sanitizedUsers[i] = ExportUser{
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
		log.Printf("couldn't write response body: %v\n", err)
		return
	}
}

func (cfg *Config) HandlerRemoveUserFromGroup(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(userContextKey).(database.User)
	if !ok {
		log.Printf("attempt to remove user from group using unauthorized user\n")
		http.Error(w, "User unauthorized", http.StatusUnauthorized)
		return
	}

	groupIDPath, ok := mux.Vars(r)["group_id"]
	if !ok {
		log.Printf("remove user from group request missing group id\n")
		http.Error(w, "Missing group id", http.StatusBadRequest)
		return
	}

	groupID, err := uuid.Parse(groupIDPath)
	if err != nil {
		log.Printf("couldn't parse group id: %v\n", err)
		http.Error(w, "Couldn't parse group id", http.StatusBadRequest)
		return
	}

	group, err := cfg.Queries.GetGroup(r.Context(), groupID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("couldn't find group: %v\n", err)
			http.Error(w, "Couldn't find group", http.StatusBadRequest)
			return
		}

		log.Printf("couldn't get group: %v\n", err)
		http.Error(w, "Somthing went wrong", http.StatusInternalServerError)
		return
	}

	if group.Owner != user.ID {
		http.Error(w, "You must be group owner to perform this action", http.StatusForbidden)
		return
	}

	data := RemoveUserFromGroupData{}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		log.Printf("couldn't decode request body: %v\n", err)
		http.Error(w, "Malformed request body", http.StatusBadRequest)
		return
	}

	if _, err := cfg.Queries.DeleteUserGroup(r.Context(), database.DeleteUserGroupParams{UserID: data.ID, GroupID: groupID}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("couln't remove user from group; user not in group: %v\n", err)
			http.Error(w, "User not in group", http.StatusBadRequest)
			return
		}

		log.Printf("couldn't delete user group: %v\n", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cfg *Config) HandlerDeleteGroup(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(userContextKey).(database.User)
	if !ok {
		log.Printf("attempt to delete group using unauthorized user\n")
		http.Error(w, "User unauthorized", http.StatusUnauthorized)
		return
	}

	groupIDPath, ok := mux.Vars(r)["group_id"]
	if !ok {
		log.Printf("delete group request missing group id\n")
		http.Error(w, "Missing group id", http.StatusBadRequest)
		return
	}

	groupID, err := uuid.Parse(groupIDPath)
	if err != nil {
		log.Printf("couldn't parse group id: %v\n", err)
		http.Error(w, "Couldn't parse group id", http.StatusBadRequest)
		return
	}

	group, err := cfg.Queries.GetGroup(r.Context(), groupID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("couldn't find group: %v\n", err)
			http.Error(w, "Couldn't find group", http.StatusBadRequest)
			return
		}

		log.Printf("coudln't get group: %v\n", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	if group.Owner != user.ID {
		log.Printf("attempt to remove user from group by non-owner\n")
		http.Error(w, "Can't remove users from group as non-owner", http.StatusForbidden)
		return
	}

	if _, err := cfg.Queries.DeleteGroup(r.Context(), groupID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("couldn't find group: %v\n", err)
			http.Error(w, "Couldn't find group", http.StatusBadRequest)
			return
		}

		log.Printf("couldn't delete group: %v\n", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
