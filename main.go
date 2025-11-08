package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/a-h/templ"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"

	"github.com/matt-horst/split-ways/handlers"
	"github.com/matt-horst/split-ways/internal/database"
	"github.com/matt-horst/split-ways/web/pages"

	_ "github.com/lib/pq"
)

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

	cfg := &handlers.Config{
		Db:     db,
		Store:  sessions.NewCookieStore([]byte(sessionKey)),
		JwtKey: jwtKey,
	}

	var _ = cfg

	router := mux.NewRouter()

	router.HandleFunc("/api/healthcheck", handlers.HandlerHealthCheck)
	router.HandleFunc("/api/users", cfg.HandlerCreateUser).Methods("POST")
	router.Handle("/api/users", cfg.AuthenticatedUserMiddleware(http.HandlerFunc(cfg.HandlerUpdateUser))).Methods("PUT")
	router.HandleFunc("/api/login", cfg.HandlerLogin).Methods("POST")
	router.Handle("/api/groups", cfg.AuthenticatedUserMiddleware(http.HandlerFunc(cfg.HandlerCreateGroup))).Methods("POST")
	router.Handle("/api/groups/{group_id}/users", cfg.AuthenticatedUserMiddleware(http.HandlerFunc(cfg.HandlerAddUserToGroup))).Methods("POST")
	router.Handle("/api/groups/{group_id}/users", cfg.AuthenticatedUserMiddleware(http.HandlerFunc(cfg.HandlerGetGroupUsers))).Methods("GET")
	router.Handle("/api/groups/{group_id}/expenses", cfg.AuthenticatedUserMiddleware(http.HandlerFunc(cfg.HandlerCreateExpense))).Methods("POST")
	router.Handle("/api/groups/{group_id}/payments", cfg.AuthenticatedUserMiddleware(http.HandlerFunc(cfg.HandlerCreatePayment))).Methods("POST")

	router.Handle("/", templ.Handler(pages.Index())).Methods("GET")
	router.Handle("/signup", templ.Handler(pages.Signup())).Methods("GET")
	router.Handle("/login", templ.Handler(pages.Login())).Methods("GET")
	router.Handle("/dashboard", cfg.AuthenticatedUserMiddleware(http.HandlerFunc(cfg.HandlerDashboard))).Methods("GET")

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static"))))

	srv := &http.Server{
		Handler:      router,
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Fatalln(err)
	}
}
