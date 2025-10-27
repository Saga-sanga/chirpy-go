package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/saga-sanga/chirpy-go/internal/auth"
	"github.com/saga-sanga/chirpy-go/internal/database"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserId    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Cannot retrieve bearer token", err)
		return
	}

	userId, err := auth.ValidateJWT(bearerToken, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Cannot validate token", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var params parameters
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Cannot decode parameters", err)
		return
	}

	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	profaneWords := []string{"kerfuffle", "sharbert", "fornax"}
	cleaned := profaneCheck(params.Body, profaneWords)

	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleaned,
		UserID: userId,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Cannot create chirp", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserId:    chirp.UserID,
	})
}

func (cfg *apiConfig) handlerRetrieveChirps(w http.ResponseWriter, r *http.Request) {
	authorIdStr := r.URL.Query().Get("author_id")
	authorId, err := uuid.Parse(authorIdStr)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "cannot parse author id", err)
		return
	}

	var dbChirps []database.Chirp
	if authorId.String() != "" {
		dbChirps, err = cfg.db.GetChirpsByAuthorID(r.Context(), authorId)
	} else {
		dbChirps, err = cfg.db.GetChirps(r.Context())
	}

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Cannot retrieve chrips from db", err)
		return
	}

	chirps := []Chirp{}

	for _, chirp := range dbChirps {
		chirps = append(chirps, Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserId:    chirp.UserID,
		})
	}

	respondWithJSON(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	strID := r.PathValue("chirpID")
	chirpUUID, err := uuid.Parse(strID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Invalid chirp id", err)
		return
	}

	dbChirp, err := cfg.db.GetChirp(r.Context(), chirpUUID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Cannot find chirp in db", err)
		return
	}

	respondWithJSON(w, http.StatusOK, Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserId:    dbChirp.UserID,
	})
}

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	s := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(s)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Invalid chirp id", err)
		return
	}

	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid bearer token", err)
		return
	}

	userID, err := auth.ValidateJWT(bearerToken, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Failed to validate token", err)
		return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Failed to fetch chirp", err)
		return
	}

	if chirp.UserID != userID {
		respondWithError(w, http.StatusForbidden, "UserID mismatch", err)
		return
	}

	_, err = cfg.db.DeleteChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}

func (cfg *apiConfig) handlerUpdateUserToChirpyRed(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Event string `json:"event"`
		Data  struct {
			UserId string `json:"user_id"`
		} `json:"data"`
	}

	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil || apiKey != cfg.polkaKey {
		respondWithError(w, http.StatusUnauthorized, "Invalid authorization header", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var req request
	if err := decoder.Decode(&req); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Cannot decode request", err)
		return
	}

	if req.Event != "user.upgraded" {
		err := fmt.Errorf("Event not recognised")
		respondWithError(w, http.StatusNoContent, "Invalid request", err)
		return
	}

	userId, err := uuid.Parse(req.Data.UserId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to parse user id", err)
		return
	}

	_, err = cfg.db.UpgradeChirpyRedByID(r.Context(), userId)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "User not found", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}
