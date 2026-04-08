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
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}

	dbQueries := database.New(db)
	config := apiConfig{db: dbQueries}
	mux := http.NewServeMux()
	mux.Handle("/app/", config.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir("./app/")))))
	mux.HandleFunc("GET /admin/healthz", handlerHealthCheck)
	mux.HandleFunc("GET /admin/metrics", config.handerMetrics)
	mux.HandleFunc("POST /admin/reset", config.handerReset)
	mux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)

	server := http.Server{Addr: ":8080", Handler: mux}
	err = server.ListenAndServe()
	if err != nil {
		fmt.Printf("failed to start a server: %+v", err)
	}
}
