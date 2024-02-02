package auth

import (
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// creates access token, straightforward
func CreateAccessToken(userID int, secretKey string, expiresInSeconds int64) (string, error) {
	// Create the Claims
	aClaims := &jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Duration(expiresInSeconds) * time.Second)),
		Subject:   strconv.Itoa(userID), // Convert userID to string
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, aClaims)
	ss, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return ss, nil
}

// creates refresh token, simple
func CreateRefreshToken(userID int, secretKey string, expiresInSeconds int64) (string, error) {
	rClaims := &jwt.RegisteredClaims{
		Issuer:    "chirpy-refresh",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Duration(expiresInSeconds) * time.Second)),
		Subject:   strconv.Itoa(userID),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, rClaims)
	ss, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return ss, nil
}
