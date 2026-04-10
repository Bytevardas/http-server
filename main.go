package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"http-server/internal/database"
)

func main() {
	godotenv.Load()

	dbURL := os.Getenv("DB_URL")
	secret := os.Getenv("JWT_SECRET")
	env := os.Getenv("PLATFORM")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}

	dbQueries := database.New(db)
	config := apiConfig{db: dbQueries, env: env, secret: secret}
	mux := http.NewServeMux()
	mux.Handle("/app/", config.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir("./app/")))))
	mux.HandleFunc("GET /admin/healthz", handlerHealthCheck)
	mux.HandleFunc("GET /admin/metrics", config.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", config.handlerReset)
	mux.HandleFunc("POST /api/chirps", config.handlerCreateChirp)
	mux.HandleFunc("GET /api/chirps", config.handlerGetAllChirps)
	mux.HandleFunc("GET /api/chirps/{chirpId}", config.handlerGetChirp)
	mux.HandleFunc("POST /api/users", config.handlerCreateUser)
	mux.HandleFunc("POST /api/login", config.handlerLogin)
	mux.HandleFunc("POST /api/refresh", config.handlerRefreshToken)
	mux.HandleFunc("POST /api/revoke", config.handlerRevokeToken)

	server := http.Server{Addr: ":8080", Handler: mux}
	err = server.ListenAndServe()
	if err != nil {
		fmt.Printf("failed to start a server: %+v", err)
	}
}
