package main

import (
	"encoding/json"
	"net/http"
	"slices"
	"sort"
	"strings"
	"time"

	"http-server/internal/auth"
	"http-server/internal/database"

	"github.com/google/uuid"
)

const maxChirpLength = 140

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

func toChirpResponses(chirps []database.Chirp) []chirpResponse {
	response := make([]chirpResponse, 0, len(chirps))
	for _, chirp := range chirps {
		response = append(response, dbChirpToResponse(chirp))
	}
	return response
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Body string `json:"body"`
	}

	defer r.Body.Close()

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "invalid bearer")
		return
	}

	user, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, 401, "invalid bearer")
		return
	}

	var params requestBody
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		respondWithError(w, 400, "invalid request body")
		return
	}

	if len(params.Body) > maxChirpLength {
		respondWithError(w, 400, "chirp exceeds maximum length of 140 characters")
		return
	}

	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		UserID: user,
		Body:   filterBadWords(params.Body),
	})
	if err != nil {
		respondWithError(w, 500, "unable to create chirp")
		return
	}

	respondWithJSON(w, 201, dbChirpToResponse(chirp))
}

func (cfg *apiConfig) handlerGetAllChirps(w http.ResponseWriter, r *http.Request) {
	sortDesc := r.URL.Query().Get("sort") == "desc"

	if authorID := r.URL.Query().Get("author_id"); authorID != "" {
		parsedID, err := uuid.Parse(authorID)
		if err != nil {
			respondWithError(w, 400, "invalid author_id")
			return
		}
		chirps, err := cfg.db.GetChirpsByAuthor(r.Context(), parsedID)
		if err != nil {
			respondWithError(w, 500, "failed to fetch chirps")
			return
		}
		sortChirps(chirps, sortDesc)
		respondWithJSON(w, 200, toChirpResponses(chirps))
		return
	}

	chirps, err := cfg.db.GetAllChirps(r.Context())
	if err != nil {
		respondWithError(w, 500, "failed to fetch chirps")
		return
	}
	sortChirps(chirps, sortDesc)
	respondWithJSON(w, 200, toChirpResponses(chirps))
}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	chirpID, err := uuid.Parse(r.PathValue("chirpId"))
	if err != nil {
		respondWithError(w, 400, "invalid chirp id")
		return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, 404, "chirp not found")
		return
	}

	respondWithJSON(w, 200, dbChirpToResponse(chirp))
}

func sortChirps(chirps []database.Chirp, desc bool) {
	sort.Slice(chirps, func(i, j int) bool {
		if desc {
			return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
		}
		return chirps[i].CreatedAt.Before(chirps[j].CreatedAt)
	})
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

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "invalid token")
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, 401, "invalid token")
		return
	}

	chirpID, err := uuid.Parse(r.PathValue("chirpId"))
	if err != nil {
		respondWithError(w, 400, "invalid chirp id")
		return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, 404, "chirp not found")
		return
	}

	if chirp.UserID != userID {
		respondWithError(w, 403, "forbidden")
		return
	}

	if err = cfg.db.DeleteChirp(r.Context(), database.DeleteChirpParams{ID: chirpID, UserID: userID}); err != nil {
		respondWithError(w, 500, "failed to delete chirp")
		return
	}

	w.WriteHeader(204)
}
