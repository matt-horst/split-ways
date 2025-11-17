package handlers

import (
	"log"
	"net/http"

	"github.com/a-h/templ"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/matt-horst/split-ways/internal/accounting"
	"github.com/matt-horst/split-ways/internal/database"
	"github.com/matt-horst/split-ways/web/pages"
)

func (cfg *Config) HandlerAddUserToGroupPage(w http.ResponseWriter, r *http.Request) {
	groupIDPath, ok := mux.Vars(r)["group_id"]
	if !ok {
		log.Printf("Couldn't find group ID in path\n")
		http.Error(w, "Couldn't find group ID", http.StatusBadRequest)
		return
	}

	groupID, err := uuid.Parse(groupIDPath)
	if err != nil {
		log.Printf("Couldn't parse group ID: %v\n", err)
		http.Error(w, "Couldn't parse group ID", http.StatusBadRequest)
		return
	}

	user, ok := r.Context().Value(userContextKey).(database.User)
	if !ok {
		log.Printf("Couldn't find authenticated user\n")
		http.Error(w, "Couldn't find authenticated user", http.StatusUnauthorized)
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
		log.Printf("Attempt to group from non member: %v\n", err)
		http.Error(w, "Not member of group", http.StatusForbidden)
		return
	}

	group, err := cfg.Db.GetGroup(r.Context(), groupID)
	if err != nil {
		log.Printf("Couldn't find group: %v\n", err)
		http.Error(w, "Couldn't find group", http.StatusBadRequest)
		return
	}

	members, err := cfg.Db.GetUsersByGroup(r.Context(), groupID)
	if err != nil {
		log.Printf("Couldn't find group members: %v\n", err)
		http.Error(w, "Couldn't find group members", http.StatusBadRequest)
		return
	}

	templ.Handler(pages.AddUser(group, user, members)).ServeHTTP(w, r)
}

func (cfg *Config) HandlerCreateExpensePage(w http.ResponseWriter, r *http.Request) {
	groupIDPath, ok := mux.Vars(r)["group_id"]
	if !ok {
		log.Printf("Couldn't find group ID in path\n")
		http.Error(w, "Couldn't find group ID", http.StatusBadRequest)
		return
	}

	groupID, err := uuid.Parse(groupIDPath)
	if err != nil {
		log.Printf("Couldn't parse group ID: %v\n", err)
		http.Error(w, "Couldn't parse group ID", http.StatusBadRequest)
		return
	}

	user, ok := r.Context().Value(userContextKey).(database.User)
	if !ok {
		log.Printf("Couldn't find authenticated user\n")
		http.Error(w, "Couldn't find authenticated user", http.StatusUnauthorized)
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
		log.Printf("Attempt to group from non member: %v\n", err)
		http.Error(w, "Not member of group", http.StatusForbidden)
		return
	}

	group, err := cfg.Db.GetGroup(r.Context(), groupID)
	if err != nil {
		log.Printf("Couldn't find group: %v\n", err)
		http.Error(w, "Couldn't find group", http.StatusBadRequest)
		return
	}

	templ.Handler(pages.CreateExpense(group)).ServeHTTP(w, r)
}

func (cfg *Config) HandlerCreatePaymentPage(w http.ResponseWriter, r *http.Request) {
	groupIDPath, ok := mux.Vars(r)["group_id"]
	if !ok {
		log.Printf("Couldn't find group ID in path\n")
		http.Error(w, "Couldn't find group ID", http.StatusBadRequest)
		return
	}

	groupID, err := uuid.Parse(groupIDPath)
	if err != nil {
		log.Printf("Couldn't parse group ID: %v\n", err)
		http.Error(w, "Couldn't parse group ID", http.StatusBadRequest)
		return
	}

	user, ok := r.Context().Value(userContextKey).(database.User)
	if !ok {
		log.Printf("Couldn't find authenticated user\n")
		http.Error(w, "Couldn't find authenticated user", http.StatusUnauthorized)
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
		log.Printf("Attempt to group from non member: %v\n", err)
		http.Error(w, "Not member of group", http.StatusForbidden)
		return
	}

	group, err := cfg.Db.GetGroup(r.Context(), groupID)
	if err != nil {
		log.Printf("Couldn't find group: %v\n", err)
		http.Error(w, "Couldn't find group", http.StatusBadRequest)
		return
	}

	templ.Handler(pages.CreatePayment(group)).ServeHTTP(w, r)
}

func (cfg *Config) HandlerGroupPage(w http.ResponseWriter, r *http.Request) {
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
		log.Printf("Attempted to view group page by unauthorized user: %v\n", err)
		http.Error(w, "User not authorized", http.StatusForbidden)
		return
	}

	txs, err := accounting.GetTransationsByGroup(cfg.Db, r.Context(), groupID)
	if err != nil {
		log.Printf("Couldn't get transactions by group: %v\n", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	group, err := cfg.Db.GetGroup(r.Context(), groupID)
	if err != nil {
		log.Printf("Couldn't find group: %v\n", err)
		http.Error(w, "Somehting went wrong", http.StatusInternalServerError)
		return
	}

	templ.Handler(pages.Group(user, group, txs)).ServeHTTP(w, r)
}
