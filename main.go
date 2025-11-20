package main

import (
	"context"
	"database/sql"
	"fmt"
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

	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "8080"
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	if err := dbConn.PingContext(ctx); err != nil {
		log.Fatalf("Couldn't establish database connection: %v\n", err)
	}
	cancel()

	cfg := &handlers.Config{
		Db:     db,
		Store:  sessions.NewCookieStore([]byte(sessionKey)),
		JwtKey: jwtKey,
	}

	router := mux.NewRouter()

	router.HandleFunc("/api/healthcheck", handlers.HandlerHealthCheck)
	router.HandleFunc("/api/users", cfg.HandlerCreateUser).Methods("POST")
	router.Handle("/api/users", cfg.AuthenticatedUserMiddleware(http.HandlerFunc(cfg.HandlerUpdateUser))).Methods("PUT")
	router.HandleFunc("/api/login", cfg.HandlerLogin).Methods("POST")
	router.HandleFunc("/api/logout", cfg.HandlerLogout).Methods("POST")
	router.HandleFunc("/api/reset", cfg.HandlerReset).Methods("POST")

	groups := router.NewRoute().PathPrefix("/api/groups").Subrouter()
	groups.Use(cfg.AuthenticatedUserMiddleware)
	groups.HandleFunc("", cfg.HandlerCreateGroup).Methods("POST")
	groups.HandleFunc("/{group_id}", cfg.HandlerUpdateGroup).Methods("PUT")
	groups.HandleFunc("/{group_id}", cfg.HandlerDeleteGroup).Methods("DELETE")
	groups.HandleFunc("/{group_id}/users", cfg.HandlerGetGroupUsers).Methods("GET")
	groups.HandleFunc("/{group_id}/users", cfg.HandlerAddUserToGroup).Methods("POST")
	groups.HandleFunc("/{group_id}/users", cfg.HandlerRemoveUserFromGroup).Methods("DELETE")
	groups.HandleFunc("/{group_id}/expenses", cfg.HandlerCreateExpense).Methods("POST")
	groups.HandleFunc("/{group_id}/expenses", cfg.HandlerUpdateExpense).Queries("id", "{id}").Methods("PUT")
	groups.HandleFunc("/{group_id}/payments", cfg.HandlerCreatePayment).Methods("POST")
	groups.HandleFunc("/{group_id}/payments", cfg.HandlerUpdatePayment).Methods("PUT")
	groups.HandleFunc("/{group_id}/transactions", cfg.HandlerDeleteTransaction).Methods("DELETE")

	router.Handle("/", cfg.AuthenticatedUserMiddleware(http.HandlerFunc(cfg.HandlerDashboard))).Methods("GET")
	router.Handle("/signup", templ.Handler(pages.Signup())).Methods("GET")
	router.Handle("/login", templ.Handler(pages.Login())).Methods("GET")
	router.Handle("/logout", templ.Handler(pages.Logout())).Methods("GET")
	router.Handle("/edit", cfg.AuthenticatedUserMiddleware(http.HandlerFunc(cfg.HandlerEditPage))).Queries("id", "{id}").Methods("GET")
	router.Handle("/groups/{group_id}", cfg.AuthenticatedUserMiddleware(http.HandlerFunc(cfg.HandlerGroupPage))).Methods("GET")
	router.Handle("/create-group", cfg.AuthenticatedUserMiddleware(templ.Handler(pages.CreateGroup())))
	router.Handle("/groups/{group_id}/manage", cfg.AuthenticatedUserMiddleware(http.HandlerFunc(cfg.HandlerManageGroupPage)))
	router.Handle("/groups/{group_id}/create-expense", cfg.AuthenticatedUserMiddleware(http.HandlerFunc(cfg.HandlerCreateExpensePage)))
	router.Handle("/groups/{group_id}/create-payment", cfg.AuthenticatedUserMiddleware(http.HandlerFunc(cfg.HandlerCreatePaymentPage)))

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static"))))

	srv := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf(":%s", port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	fmt.Printf("Now listening on %s!\n", srv.Addr)

	err = srv.ListenAndServe()
	if err != nil {
		log.Fatalln(err)
	}
}
