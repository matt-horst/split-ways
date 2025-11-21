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

	"github.com/matt-horst/split-ways/handlers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"database/sql"

	"github.com/gorilla/sessions"

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

	assert.Equal(t, rr.Code, http.StatusNoContent)

	// Test creating user with taken username responds with error
	r = httptest.NewRequest("POST", "/api/users", bytes.NewBuffer(body))
	rr = httptest.NewRecorder()

	cfg.HandlerCreateUser(rr, r)

	assert.Equal(t, rr.Code, http.StatusBadRequest)
}
