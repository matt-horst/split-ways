package handlers

import (
	"github.com/gorilla/sessions"
	"github.com/matt-horst/split-ways/internal/database"
)

type Config struct {
	Db     *database.Queries
	Store  *sessions.CookieStore
	JwtKey string
}
