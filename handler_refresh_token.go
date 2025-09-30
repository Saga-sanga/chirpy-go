package main

import (
	"net/http"
	"time"

	"github.com/saga-sanga/chirpy-go/internal/auth"
)

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}
	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "No bearer token found", err)
		return
	}

	refreshToken, err := cfg.db.GetRefreshToken(r.Context(), bearerToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Failed to find token in db", err)
		return
	}

	expired := time.Now().Compare(refreshToken.ExpiresAt)
	if expired == 1 {
		respondWithError(w, http.StatusUnauthorized, "Expired refresh token", err)
		return
	}

	user, err := cfg.db.GetUserFromRefreshToken(r.Context(), bearerToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch user from token", err)
		return
	}

	tokenExpiration := time.Hour
	accessToken, err := auth.MakeJWT(user.ID, cfg.secret, tokenExpiration)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to make access token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		Token: accessToken,
	})
}
