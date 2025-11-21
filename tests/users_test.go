package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"database/sql"

	"github.com/gorilla/sessions"

	"github.com/matt-horst/split-ways/handlers"
	"github.com/matt-horst/split-ways/internal/api"
	"github.com/matt-horst/split-ways/internal/database"

	_ "github.com/lib/pq"
)

var db *sql.DB
var queries *database.Queries

const sessionKey string = "session_key"
const jwtKey string = "jwt_key"

func TestMain(m *testing.M) {
	var err error
	db, err = sql.Open("postgres", "postgresql://postgres:postgres@localhost:5432/split_ways_test?sslmode=disable")
	if err != nil {
		log.Fatalf("could not connect to database: %v\n", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("could not ping database: %v\n", err)
	}

	queries = database.New(db)

	exitCode := m.Run()

	if err := api.Reset(context.Background(), db, queries); err != nil {
		log.Fatalf("could not clean up test database: %v\n", err)
	}

	os.Exit(exitCode)
}

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

	assert.Equal(t, http.StatusNoContent, rr.Code)

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

	require.Equal(t, http.StatusNoContent, rr.Code)

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

	require.Equal(t, http.StatusNoContent, rr.Code)

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
