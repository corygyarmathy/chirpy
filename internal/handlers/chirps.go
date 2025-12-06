package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"

	"github.com/corygyarmathy/chirpy/internal/auth"
	"github.com/corygyarmathy/chirpy/internal/database"
	"github.com/google/uuid"
)

func (api *API) GetChirps(w http.ResponseWriter, r *http.Request) {
	var chirps []database.Chirp
	var err error

	s := r.URL.Query().Get("author_id")

	if s != "" {
		authorUUID, err := uuid.Parse(s)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "GetChirps: couldn't parse path value 'author_id' to UUID", err)
		}

		chirps, err = api.DB.GetChirpsByUserID(r.Context(), authorUUID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "GetChirps: couldn't get chirps from DB", err)
		}
	} else {
		chirps, err = api.DB.GetChirps(r.Context())
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "GetChirps: couldn't get chirps from DB", err)
		}
	}

	respondWithJSON(w, http.StatusOK, chirps)
}

func (api *API) GetChirpByID(w http.ResponseWriter, r *http.Request) {
	chirpID := r.PathValue("chirpID")
	id, err := uuid.Parse(chirpID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "GetChirps: couldn't parse path value 'chirpID' to UUID", err)
	}
	chirp, err := api.DB.GetChirpByID(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "GetChirps: no chirps found for the given ID", err)
		}
		respondWithError(w, http.StatusInternalServerError, "GetChirps: couldn't get chirps from DB", err)
	}
	respondWithJSON(w, http.StatusOK, chirp)
}

func (api *API) CreateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	var params parameters
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "CreateChirp: couldn't decode parameters", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "CreateChirp: failed to get bearer token", err)
		return
	}

	userUUID, err := auth.ValidateJWT(token, api.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "CreateChirp: couldn't validate user JWT", err)
		return
	}

	cBody, err := validateChirp(params.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "CreateChirp: couldn't validate chirp", err)
		return
	}

	chirp, err := api.DB.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cBody,
		UserID: userUUID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "CreateChirp: couldn't create chirp in DB", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, chirp)
}

func validateChirp(body string) (string, error) {
	const maxChirpLength = 140
	if len(body) > maxChirpLength {
		return "", errors.New("chirp is too long")
	}

	cBody := profanityCensor(body)

	return cBody, nil
}

func profanityCensor(s string) string {
	profanity := []string{"kerfuffle", "sharbert", "fornax"}

	words := strings.Split(s, " ")
	for i, word := range words {
		word = strings.ToLower(word)
		if slices.Contains(profanity, word) {
			words[i] = "****"
		}
	}

	return strings.Join(words, " ")
}

func (api *API) DeleteChirp(w http.ResponseWriter, r *http.Request) {
	chirpID := r.PathValue("chirpID")
	chirpUUID, err := uuid.Parse(chirpID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "DeleteChirp: couldn't parse path value 'chirpID' to UUID", err)
	}
	chirp, err := api.DB.GetChirpByID(r.Context(), chirpUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "DeleteChirp: no chirps found for the given ID", err)
		}
		respondWithError(w, http.StatusInternalServerError, "DeleteChirp: couldn't get chirps from DB", err)
	}

	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "DeleteChirp: failed to get access token from request header", err)
		return
	}

	userID, err := auth.ValidateJWT(accessToken, api.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "DeleteChirp: user JWT not authorised", err)
		return
	}

	if chirp.UserID != userID {
		respondWithError(w, http.StatusForbidden, "DeleteChirp: user ID does not match chirp user ID", err)
		return
	}

	if err = api.DB.DeleteChirp(r.Context(), chirp.ID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "DeleteChirp: failed to delete chirp", err)
		return
	}
	respondWithJSON(w, http.StatusNoContent, nil)
}
