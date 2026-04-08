package main

import (
	"fmt"
	"net/http"
)

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, req *http.Request) {
	if cfg.env != "dev" {
		respondWithError(w, 403, "can't reset the database on non dev environment")
	}

	if err := cfg.db.DeleteAllUsers(req.Context()); err != nil {
		respondWithError(w, 500, "failed to delete users")
	}

	respondWithJSON(w, 200, "success")
}
