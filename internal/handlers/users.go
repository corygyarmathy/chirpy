package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/corygyarmathy/chirpy/internal/auth"
	"github.com/corygyarmathy/chirpy/internal/database"
	"github.com/google/uuid"
)

func (api *API) CreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var params parameters
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "CreateUser: couldn't decode parameters", err)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "CreateUser: couldn't hash password", err)
		return
	}

	user, err := api.DB.CreateUser(r.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "CreateUser: couldn't create user in DB", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, user)
}

func (api *API) LoginUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds *int   `json:"expires_in_seconds,omitempty"`
	}

	type LoggedInUser struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
		Token     string    `json:"token"`
	}

	var params parameters
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "LoginUser: couldn't decode parameters", err)
		return
	}

	if params.ExpiresInSeconds == nil || *params.ExpiresInSeconds > 3600 {
		defaultExpiry := 3600
		params.ExpiresInSeconds = &defaultExpiry
	}

	user, err := api.DB.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	match, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)

	if !match || err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	token, err := auth.MakeJWT(user.ID, api.jwtSecret, time.Duration(*params.ExpiresInSeconds)*time.Second)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "LoginUser: failed to get bearer token", err)
		return
	}

	loggedInUser := LoggedInUser{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		Token:     token,
	}

	respondWithJSON(w, http.StatusOK, loggedInUser)
}
