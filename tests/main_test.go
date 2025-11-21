package tests

import (
	"context"
	"database/sql"
	"log"
	"os"
	"testing"

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
