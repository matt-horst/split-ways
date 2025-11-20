package handlers

import (
	"fmt"
	"log"
	"net/http"
)

func (cfg *Config) HandlerReset(w http.ResponseWriter, r *http.Request) {
	if err := cfg.Db.DeleteAllGroups(r.Context()); err != nil {
		log.Printf("Couldn't delete groups: %v\n", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	if err := cfg.Db.DeleteAllUsers(r.Context()); err != nil {
		log.Printf("Couldn't delete users: %v\n", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	fmt.Printf("Successfully reset database\n")

	w.WriteHeader(http.StatusNoContent)
}
