package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/matt-horst/split-ways/internal/auth"
	"github.com/matt-horst/split-ways/internal/database"
	"github.com/matt-horst/split-ways/web/pages"
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
		log.Printf("Couldn't create new user: %v\n", err)
		http.Error(w, "Username taken", http.StatusBadRequest)
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.JwtKey, time.Hour)
	if err != nil {
		log.Printf("Couldn't create JWT: %v\n", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	session, err := cfg.Store.Get(r, "user-session")
	if err != nil {
		log.Printf("Couldn't get session: %v\n", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	session.Values["token"] = token
	err = session.Save(r, w)
	if err != nil {
		log.Printf("Couldn't save session:L %v\n", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
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

func (cfg *Config) HandlerUpdateUser(userID uuid.UUID, w http.ResponseWriter, r *http.Request) {
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

	_, err = cfg.Db.UpdatePassword(r.Context(), database.UpdatePasswordParams{
		ID:             userID,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		log.Printf("Couldn't update password: %v\n", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cfg *Config) HandlerDashboard(userID uuid.UUID, w http.ResponseWriter, r *http.Request) {
	user, err := cfg.Db.GetUserByID(r.Context(), userID)
	if err != nil {
		log.Printf("Couldn't find user: %v\n", err)
		http.Error(w, "Couldn't find user", http.StatusBadRequest)
		return
	}

	err = pages.Dashboard(user.Username).Render(r.Context(), w)
	if err != nil {
		log.Printf("Couldn't send page: %v\n", err)
		return
	}
}

func (cfg *Config) UserMiddleware(fn func(uuid.UUID, http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		fn(userID, w, r)
	}
}
