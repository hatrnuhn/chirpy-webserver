package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func (cfg *apiConfig) handlePostRevoke(w http.ResponseWriter, r *http.Request) {
	// reads rToken from Header
	authHead := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHead, "Bearer ") {
		respondWithError(w, 401, "invalid auth header")
		return
	}

	rTokenString := strings.TrimPrefix(authHead, "Bearer ")

	// parses rTokenString
	rToken, err := jwt.ParseWithClaims(rTokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(cfg.jwtSecret), nil
	})

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	if claims, ok := rToken.Claims.(*jwt.RegisteredClaims); ok {
		// verifs RJWT
		if claims.Issuer != "chirpy-refresh" {
			respondWithError(w, http.StatusUnauthorized, "invalid RJWT")
			return
		}

		// verifs RJWT expiration
		isNotExp, err := cfg.db.RJWTNotExp(rTokenString)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("couldn't verify RJWT expiration status: %s", err.Error()))
			return
		}

		if !isNotExp {
			respondWithError(w, http.StatusUnauthorized, "RJWT is expired")
			return
		}

		// updates tokens on db
		_, err = cfg.db.WriteRefreshToken(rTokenString, time.Now().Unix())
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("couldn't write RJWT to db: %s", err.Error()))
			return
		}

		// returns with StatusOK
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	} else {
		respondWithError(w, http.StatusBadRequest, "invalid RJWT")
	}

}
