package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"

	"github.com/corygyarmathy/chirpy/internal/database"
	"github.com/google/uuid"
)

func (api *API) CreateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	var params parameters
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "CreateChirp: couldn't decode parameters", err)
		return
	}

	cBody, err := validateChirp(params.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "CreateChirp: couldn't validate chirp", err)
		return
	}

	chirp, err := api.DB.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cBody,
		UserID: params.UserID,
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
