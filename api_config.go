package main

import (
	"net/http"
	"sync/atomic"

	"http-server/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	env            string
	db             *database.Queries
	secret         string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}
