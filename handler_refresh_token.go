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
		respondWithError(w, http.StatusBadRequest, "No bearer token found", err)
		return
	}

	user, err := cfg.db.GetUserFromRefreshToken(r.Context(), bearerToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Failed to fetch user from token", err)
		return
	}

	accessToken, err := auth.MakeJWT(user.ID, cfg.secret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to make access token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		Token: accessToken,
	})
}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Bearer token not found", err)
		return
	}

	_, err = cfg.db.RevokeRefreshToken(r.Context(), bearerToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to revoke refresh token", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
