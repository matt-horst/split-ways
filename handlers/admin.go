package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/matt-horst/split-ways/internal/api"
)

func (cfg *Config) HandlerReset(w http.ResponseWriter, r *http.Request) {
	if err := api.Reset(r.Context(), cfg.DB, cfg.Queries); err != nil {
		log.Printf("couldn't reset database: %v\n", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	fmt.Printf("Successfully reset database\n")

	w.WriteHeader(http.StatusNoContent)
}
