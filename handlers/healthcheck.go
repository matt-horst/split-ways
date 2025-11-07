package handlers

import (
	"net/http"
)

func HandlerHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}
