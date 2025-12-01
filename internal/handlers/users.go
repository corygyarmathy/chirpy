package handlers

import (
	"encoding/json"
	"net/http"
)

func (api *API) CreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
	}

	var params parameters
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "CreateUser: couldn't decode parameters", err)
		return
	}

	user, err := api.DB.CreateUser(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "CreateUser: couldn't create user in DB", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, user)
}
