package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"

	"github.com/matt-horst/split-ways/internal/database"

	_ "github.com/lib/pq"
)

type config struct {
	db     *database.Queries
	store  *sessions.CookieStore
	jwtKey string
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Couldn't load ENV file: %v\n", err)
	}

	dbConnStr, ok := os.LookupEnv("DATABASE")
	if !ok {
		log.Fatalln("Couldn't find database connection string in ENV")
	}

	sessionKey, ok := os.LookupEnv("SESSION_KEY")
	if !ok {
		log.Fatalln("Couldn't find session key in ENV")
	}

	jwtKey, ok := os.LookupEnv("JWT_KEY")
	if !ok {
		log.Fatalln("Couldn't find jwt key in ENV")
	}

	dbConn, err := sql.Open("postgres", dbConnStr)
	if err != nil {
		log.Fatalf("Couldn't open database connection: %v\n", err)
	}

	db := database.New(dbConn)

	cfg := &config{
		db:     db,
		store:  sessions.NewCookieStore([]byte(sessionKey)),
		jwtKey: jwtKey,
	}

	var _ = cfg

	router := mux.NewRouter()

	router.HandleFunc("/api/healthcheck", handleHealthCheck)

	srv := &http.Server{
		Handler:      router,
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}
