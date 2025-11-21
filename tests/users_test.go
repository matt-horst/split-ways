package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorilla/sessions"

	"github.com/matt-horst/split-ways/handlers"
)

func TestCreateUser(t *testing.T) {
	tx, err := db.Begin()
	require.NoError(t, err)
	defer tx.Rollback()

	cfg := &handlers.Config{
		DB:      db,
		Queries: queries.WithTx(tx),
		Store:   sessions.NewCookieStore([]byte(sessionKey)),
		JwtKey:  jwtKey,
	}
	// Test successful creation of user
	type createUserData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	data := createUserData{
		Username: "user",
		Password: "password",
	}

	body, err := json.Marshal(data)
	require.NoError(t, err)

	r := httptest.NewRequest("POST", "/api/users", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	cfg.HandlerCreateUser(rr, r)

	assert.Equal(t, http.StatusCreated, rr.Code)

	// Test creating user with taken username responds with error
	r = httptest.NewRequest("POST", "/api/users", bytes.NewBuffer(body))
	rr = httptest.NewRecorder()

	cfg.HandlerCreateUser(rr, r)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandlerLogin(t *testing.T) {
	// Setup
	tx, err := db.Begin()
	require.NoError(t, err)
	defer tx.Rollback()

	cfg := &handlers.Config{
		DB:      db,
		Queries: queries.WithTx(tx),
		Store:   sessions.NewCookieStore([]byte(sessionKey)),
		JwtKey:  jwtKey,
	}

	type userData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	data := userData{
		Username: "user",
		Password: "password",
	}

	body, err := json.Marshal(data)
	require.NoError(t, err)

	r := httptest.NewRequest("POST", "/api/users", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	cfg.HandlerCreateUser(rr, r)

	require.Equal(t, http.StatusCreated, rr.Code)

	// Test successful login
	r = httptest.NewRequest("POST", "/api/login", bytes.NewBuffer(body))
	rr = httptest.NewRecorder()

	cfg.HandlerLogin(rr, r)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	cookies := rr.Result().Cookies()
	assert.NotEmpty(t, cookies)

	// Test login with wrong password
	data = userData{
		Username: "user",
		Password: "incorrect_password",
	}

	body, err = json.Marshal(data)
	assert.NoError(t, err)

	r = httptest.NewRequest("POST", "/api/login", bytes.NewBuffer(body))
	rr = httptest.NewRecorder()

	cfg.HandlerLogin(rr, r)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	// Test login with wrong username
	data = userData{
		Username: "incorrect_user",
		Password: "password",
	}

	body, err = json.Marshal(data)
	assert.NoError(t, err)

	r = httptest.NewRequest("POST", "/api/login", bytes.NewBuffer(body))
	rr = httptest.NewRecorder()

	cfg.HandlerLogin(rr, r)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestUpdateUser(t *testing.T) {
	// Setup
	tx, err := db.Begin()
	require.NoError(t, err)
	defer tx.Rollback()

	cfg := &handlers.Config{
		DB:      db,
		Queries: queries.WithTx(tx),
		Store:   sessions.NewCookieStore([]byte(sessionKey)),
		JwtKey:  jwtKey,
	}

	type userData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	data := userData{
		Username: "user",
		Password: "password",
	}

	body, err := json.Marshal(data)
	require.NoError(t, err)

	r := httptest.NewRequest("POST", "/api/users", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	cfg.HandlerCreateUser(rr, r)

	require.Equal(t, http.StatusCreated, rr.Code)

	r = httptest.NewRequest("POST", "/api/login", bytes.NewBuffer(body))
	rr = httptest.NewRecorder()

	cfg.HandlerLogin(rr, r)

	cookies := rr.Result().Cookies()

	// Test successful update
	type updateUserData struct {
		Password string `json:"password"`
	}

	body, err = json.Marshal(updateUserData{Password: "new_password"})
	assert.NoError(t, err)

	r = httptest.NewRequest("PUT", "/api/users", bytes.NewBuffer(body))
	for _, c := range cookies {
		r.AddCookie(c)
	}

	rr = httptest.NewRecorder()

	cfg.AuthenticatedUserMiddleware(http.HandlerFunc(cfg.HandlerUpdateUser)).ServeHTTP(rr, r)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}
