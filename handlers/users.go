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

	user, err := cfg.Queries.CreateUser(
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

	user, err := cfg.Queries.GetUserByUsername(r.Context(), data.Username)
	if err != nil {
		log.Printf("Couldn't find user: %v\n", err)
		http.Error(w, "Couldn't find user", http.StatusUnauthorized)
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

func (cfg *Config) HandlerLogout(w http.ResponseWriter, r *http.Request) {
	session, err := cfg.Store.Get(r, "user-session")
	if err != nil {
		log.Printf("Couldn't get cookie store: %v\n", err)
		http.Error(w, "Something went wrong", http.StatusBadRequest)
		return
	}

	session.Options.MaxAge = -1
	if err := session.Save(r, w); err != nil {
		log.Printf("Couldn't revoke user-session cookie: %v\n", err)
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

	_, err = cfg.Queries.UpdatePassword(r.Context(), database.UpdatePasswordParams{
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

		user, err := cfg.Queries.GetUserByID(r.Context(), userID)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, user)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
