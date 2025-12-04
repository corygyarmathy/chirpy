package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	issuer = "chirpy"
)

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    issuer,
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject:   userID.String(),
	})

	signingKey := []byte(tokenSecret)
	signed, err := token.SignedString(signingKey)
	if err != nil {
		return "", fmt.Errorf("signing token string with secret: %v", err)
	}

	return signed, nil
}

func ValidateJWT(tokenString string, tokenSecret string) (uuid.UUID, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(t *jwt.Token) (any, error) {
			return []byte(tokenSecret), nil
		},
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("parsing token with claims: %v", err)
	}

	userID, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, fmt.Errorf("getting token claims subject: %v", err)
	}

	cIssuer, err := token.Claims.GetIssuer()
	if err != nil {
		return uuid.Nil, err
	}
	if cIssuer != string(issuer) {
		return uuid.Nil, errors.New("invalid issuer")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("converting string %v to UUID: %v", userID, err)
	}

	return userUUID, nil
}
