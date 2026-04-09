package main

import (
	"encoding/json"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	"http-server/internal/database"

	"github.com/google/uuid"
)

var badWords = []string{"kerfuffle", "sharbert", "fornax"}

type chirpResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    uuid.UUID `json:"user_id"`
	Body      string    `json:"body"`
}

func dbChirpToResponse(chirp database.Chirp) chirpResponse {
	return chirpResponse{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		UserID:    chirp.UserID,
		Body:      chirp.Body,
	}
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, req *http.Request) {
	type requestBody struct {
		Body   string    `json:"body"`
		UserId uuid.UUID `json:"user_id"`
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

	chirp, err := cfg.db.CreateChirp(req.Context(), database.CreateChirpParams{
		UserID: params.UserId,
		Body:   filterBadWords(params.Body),
	})
	if err != nil {
		respondWithError(w, 500, "unable to create chirp")
		return
	}

	respondWithJSON(w, 201, dbChirpToResponse(chirp))
}

func (cfg *apiConfig) handlerGetAllChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.db.GetAllChirps(r.Context())
	if err != nil {
		respondWithError(w, 500, "failed to fetch chirps")
		return
	}

	response := make([]chirpResponse, 0, len(chirps))
	for _, chirp := range chirps {
		response = append(response, dbChirpToResponse(chirp))
	}

	respondWithJSON(w, 200, response)
}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	value := r.PathValue("chirpId")
	chirpId, err := uuid.Parse(value)
	if err != nil {
		respondWithError(w, 400, "value is not uuid")
		return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), chirpId)
	if err != nil {
		respondWithError(w, 404, "chirp not found")
		return
	}

	respondWithJSON(w, 200, dbChirpToResponse(chirp))
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
