package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"sync/atomic"
)

var badWords = []string{"kerfuffle", "sharbert", "fornax"}

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
	mux.HandleFunc("GET /admin/healthz", handlerHealthCheck)
	mux.HandleFunc("GET /admin/metrics", config.handerMetrics)
	mux.HandleFunc("POST /admin/reset", config.handerReset)
	mux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)

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
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) handerReset(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	cfg.fileserverHits.Swap(0)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("200 OK"))
}

func handlerValidateChirp(w http.ResponseWriter, req *http.Request) {
	type requestBody struct {
		Body string `json:"body"`
	}

	defer req.Body.Close()
	data, err := io.ReadAll(req.Body)
	if err != nil {
		respondWithError(w, 500, "could not read the body")
		return
	}

	params := requestBody{}
	err = json.Unmarshal(data, &params)
	if err != nil {
		respondWithError(w, 500, "could not unmarshal the body")
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, 400, "chirp is longer the 140 chars")
		return
	}

	filteredSentence := filterBadWords(params.Body)
	type responseBody struct {
		Valid       bool   `json:"valid"`
		CleanedBody string `json:"cleaned_body,omitempty"`
	}
	respondWithJSON(w, 200, responseBody{Valid: true, CleanedBody: filteredSentence})
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	respondWithJSON(w, code, map[string]string{"error": msg})
}

func respondWithJSON(w http.ResponseWriter, code int, payload any) error {
	response, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
	return nil
}

func filterBadWords(sentence string) string {
	words := strings.Split(sentence, " ")
	filtered := make([]string, 0, len(words))
	for _, word := range words {
		if slices.Contains(badWords, strings.ToLower(word)) {
			filtered = append(filtered, "****")
			continue
		}
		filtered = append(filtered, word)
	}
	return strings.Join(filtered, " ")
}
