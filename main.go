package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)

		next.ServeHTTP(w, r)
	})
}

func main() {
	config := apiConfig{}
	mux := http.NewServeMux()
	mux.Handle("/app/", config.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir("./app/")))))
	mux.HandleFunc("/healthz", handlerHealthCheck)
	mux.HandleFunc("/metrics", config.handerMetrics)
	mux.HandleFunc("/reset", config.handerReset)

	server := http.Server{Addr: ":8080", Handler: mux}
	err := server.ListenAndServe()
	if err != nil {
		fmt.Printf("failed to start a server: %+v", err)
	}
}

func handlerHealthCheck(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("200 OK"))
}

func (cfg *apiConfig) handerMetrics(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(fmt.Append([]byte{}, "Hits: ", cfg.fileserverHits.Load()))
}

func (cfg *apiConfig) handerReset(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	cfg.fileserverHits.Swap(0)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("200 OK"))
}
