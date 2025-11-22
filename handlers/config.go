package handlers

import (
	"database/sql"
	"github.com/gorilla/sessions"
	"github.com/matt-horst/split-ways/internal/database"
)

type Config struct {
	DB      *sql.DB
	Tx      *sql.Tx
	Queries *database.Queries
	Store   *sessions.CookieStore
	JwtKey  string
}
