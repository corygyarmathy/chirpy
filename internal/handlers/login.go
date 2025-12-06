package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/corygyarmathy/chirpy/internal/auth"
	"github.com/corygyarmathy/chirpy/internal/database"
	"github.com/google/uuid"
)

func (api *API) LoginUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type LoggedInUser struct {
		ID           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
		IsChirpyRed  bool      `json:"is_chirpy_red"`
	}

	var params parameters
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "LoginUser: couldn't decode parameters", err)
		return
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

	accessToken, err := auth.MakeJWT(user.ID, api.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "LoginUser: couldn't create access JWT", err)
		return
	}

	refreshTokenString, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "LoginUser: failed to make refresh token", err)
		return
	}

	refreshToken, err := api.DB.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshTokenString,
		ExpiresAt: time.Now().UTC().UTC().Add(60 * 24 * time.Hour),
		UserID:    user.ID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "LoginUser: failed to store refresh token in DB", err)
		return
	}

	loggedInUser := LoggedInUser{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        accessToken,
		RefreshToken: refreshToken.Token,
		IsChirpyRed:  user.IsChirpyRed,
	}

	respondWithJSON(w, http.StatusOK, loggedInUser)
}

func (api *API) RefreshLogin(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}

	refreshTokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "RefreshLogin: failed to get bearer token from request header", err)
		return
	}

	refreshToken, err := api.DB.GetRefreshTokenByToken(r.Context(), refreshTokenString)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusUnauthorized, "RefreshLogin: no refresh token found for the given token string", err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "RefreshLogin: couldn't get refreshToken from DB", err)
		return
	}

	user, err := api.DB.GetUserFromValidRefreshToken(r.Context(), refreshToken.Token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "RefreshLogin: couldn't get user for token from DB", err)
		return
	}

	accessToken, err := auth.MakeJWT(user.ID, api.jwtSecret, 1*time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "RefreshLogin: couldn't make new JWT for user", err)
		return
	}

	newToken := response{accessToken}
	respondWithJSON(w, http.StatusOK, newToken)
}

func (api *API) RevokeLogin(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "RevokeLogin: failed to get bearer token from request header", err)
		return
	}

	refreshToken, err := api.DB.GetRefreshTokenByToken(r.Context(), token)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "RevokeLogin: no refresh token found for the given token string", err)
		}
		respondWithError(w, http.StatusInternalServerError, "RevokeLogin: couldn't get refreshToken from DB", err)
	}

	err = api.DB.RevokeRefreshToken(r.Context(), refreshToken.UserID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "RevokeLogin: failed revoke refresh login token in DB", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}
