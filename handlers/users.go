package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/matt-horst/split-ways/internal/auth"
	"github.com/matt-horst/split-ways/internal/database"
	"github.com/matt-horst/split-ways/web/pages"
)

type contextKey string

const (
	userContextKey contextKey = "user"
)

func (cfg *Config) HandlerCreateUser(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Printf("Couldn't decode body: %v\n", err)
		http.Error(w, "Couldn't create new user", http.StatusBadRequest)
		return
	}

	hashedPassword, err := auth.HashPassword(data.Password)
	if err != nil {
		log.Printf("Couldn't hash password: %v\n", err)
		http.Error(w, "Couldn't create new user", http.StatusBadRequest)
		return
	}

	user, err := cfg.Db.CreateUser(
		r.Context(),
		database.CreateUserParams{
			Username:       data.Username,
			HashedPassword: hashedPassword,
		},
	)
	if err != nil {
		if strings.Contains(err.Error(), "unique") && strings.Contains(err.Error(), "username") {
			log.Printf("Couldn't create new user: %v\n", err)
			http.Error(w, "Username taken", http.StatusBadRequest)
			return
		}

		log.Printf("Couldn't create new user: %v\n", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.JwtKey, time.Hour)
	if err != nil {
		log.Printf("Couldn't create JWT: %v\n", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	err = auth.SetBearerToken(cfg.Store, w, r, token)
	if err != nil {
		log.Printf("Couldn't set bearer token: %v\n", err)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cfg *Config) HandlerLogin(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Printf("Couldn't decode request body: %v", err)
		http.Error(w, "Couldn't login", http.StatusBadRequest)
		return
	}

	user, err := cfg.Db.GetUserByUsername(r.Context(), data.Username)
	if err != nil {
		log.Printf("Couldn't find user: %v\n", err)
		http.Error(w, "Couldn't find user", http.StatusBadRequest)
		return
	}

	ok, err := auth.CheckPasswordHash(data.Password, user.HashedPassword)
	if err != nil {
		log.Printf("Couldn't check password hash: %v\n", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(w, "Username and password do not match", http.StatusUnauthorized)
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.JwtKey, time.Hour)
	if err != nil {
		log.Printf("Couldn't create new JWT: %v\n", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	err = auth.SetBearerToken(cfg.Store, w, r, token)
	if err != nil {
		log.Printf("Couldn't set the bearer token: %v\n", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cfg *Config) HandlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Password string `json:"password"`
	}{}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Printf("Couldn't decode request body: %v\n", err)
		http.Error(w, "Couldn't change passowrd", http.StatusBadRequest)
		return
	}

	hashedPassword, err := auth.HashPassword(data.Password)
	if err != nil {
		log.Printf("Couldn't hash password: %v\n", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	user, ok := r.Context().Value(userContextKey).(database.User)
	if !ok {
		log.Printf("Attempted to update unauthenticated user\n")
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	_, err = cfg.Db.UpdatePassword(r.Context(), database.UpdatePasswordParams{
		ID:             user.ID,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		log.Printf("Couldn't update password: %v\n", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cfg *Config) HandlerDashboard(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(userContextKey).(database.User)
	if !ok {
		log.Printf("Attempted to handle dashboard with unauthenticated user\n")
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	groups, err := cfg.Db.GetGroupsByUser(r.Context(), user.ID)
	if err != nil {
		log.Printf("Couldn't find groups by user: %v\n", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	err = pages.Dashboard(user.Username, groups).Render(r.Context(), w)
	if err != nil {
		log.Printf("Couldn't send page: %v\n", err)
		return
	}
}

func (cfg *Config) AuthenticatedUserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.GetBearerToken(cfg.Store, r)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		userID, err := auth.ValidateJWT(token, cfg.JwtKey)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		user, err := cfg.Db.GetUserByID(r.Context(), userID)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, user)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
