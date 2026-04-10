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

const (
	accessTokenTTL  = time.Hour
	refreshTokenTTL = 60 * 24 * time.Hour
)

type userResponse struct {
	ID           uuid.UUID `json:"id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email,omitempty"`
	Token        string    `json:"token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
}

type requestBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func dbUserToResponse(user database.User, token string, refreshToken string) userResponse {
	return userResponse{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refreshToken,
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

	respondWithJSON(w, 201, dbUserToResponse(user, "", ""))
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

	token, err := auth.MakeJWT(user.ID, cfg.secret, accessTokenTTL)
	if err != nil {
		respondWithError(w, 500, "failed to generate token")
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, 500, "failed to generate refresh token")
		return
	}

	_, err = cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().UTC().Add(refreshTokenTTL),
	})
	if err != nil {
		respondWithError(w, 500, "failed to save refreshToken")
		return
	}

	respondWithJSON(w, 200, dbUserToResponse(user, token, refreshToken))
}

func (cfg *apiConfig) handlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "invalid token")
		return
	}

	refreshToken, err := cfg.db.GetRefreshToken(r.Context(), token)
	if err != nil {
		respondWithError(w, 401, "token not found, expired, or revoked")
		return
	}

	accessToken, err := auth.MakeJWT(refreshToken.UserID, cfg.secret, accessTokenTTL)
	if err != nil {
		respondWithError(w, 500, "failed to generate token")
		return
	}

	respondWithJSON(w, 200, userResponse{Token: accessToken})
}

func (cfg *apiConfig) handlerRevokeToken(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "invalid token")
		return
	}

	result, err := cfg.db.RevokeRefreshToken(r.Context(), token)
	if err != nil {
		respondWithError(w, 500, "failed to revoke token")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		respondWithError(w, 500, "failed to revoke token")
		return
	}
	if rowsAffected == 0 {
		respondWithError(w, 404, "token not found")
		return
	}

	w.WriteHeader(204)
}
