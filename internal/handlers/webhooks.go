package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/corygyarmathy/chirpy/internal/auth"
	"github.com/google/uuid"
)

func (api *API) Webhooks(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	var params parameters
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Webhooks: couldn't decode parameters", err)
		return
	}

	if params.Event != "user.upgraded" {
		respondWithJSON(w, http.StatusNoContent, nil)
	}

	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Webhooks: failed to get API key from request heaeder", err)
		return
	}

	if apiKey != api.polkaKey {
		respondWithError(w, http.StatusUnauthorized, "Webhooks: provided API key does not match", nil)
		return
	}

	_, err = api.DB.SetChirpyRedActive(r.Context(), params.Data.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Webhooks: no user found for given ID", err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Webhooks: failed to set chirpy red to active in DB", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}
