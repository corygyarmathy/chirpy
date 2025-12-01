package handlers

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"
)

func (api *API) ValidateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	type response struct {
		CleanedBody string `json:"cleaned_body"`
	}

	var params parameters
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	cleaned := profanityCensor(params.Body)

	respondWithJSON(w, http.StatusOK, response{
		CleanedBody: cleaned,
	})
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
