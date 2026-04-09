package main

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"

	"http-server/internal/auth"
	"http-server/internal/database"
)

type userResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type requestBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
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

	hash, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, 500, "failed to hash password")
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{Email: params.Email, HashedPassword: hash})
	if err != nil {
		respondWithError(w, 500, "failed to create the user")
		return
	}

	respondWithJSON(w, 201, dbUserToResponse(user))
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	data, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, 500, "failed to read the body")
		return
	}

	var params requestBody
	if err = json.Unmarshal(data, &params); err != nil {
		respondWithError(w, 500, "failed to unmarshal params")
		return
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, 404, "user not found")
		return
	}

	match, err := auth.CheckPassword(params.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, 500, "failed to check password")
		return
	}

	if !match {
		respondWithError(w, 401, "incorrect password")
		return
	}

	respondWithJSON(w, 200, dbUserToResponse(user))
}
