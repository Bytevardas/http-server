package main

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"

	"http-server/internal/database"
)

type userResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func dbUserToResponse(user database.User) userResponse {
	return userResponse{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	type requestBody struct {
		Email string `json:"email"`
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, 500, "failed to read the body")
		return
	}

	var params requestBody
	if err = json.Unmarshal(data, &params); err != nil {
		respondWithError(w, 500, "failed to unmarshal data")
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, 500, "failed to create the user")
		return
	}

	respondWithJSON(w, 201, dbUserToResponse(user))
}
