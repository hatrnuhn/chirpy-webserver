package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hatrnuhn/chirpy-webserver/internal/auth"
)

// refreshes access token using refresh token
func (cfg *apiConfig) handlePostRefresh(w http.ResponseWriter, r *http.Request) {
	rToken, err := auth.ParseReq(r, cfg.jwtSecret, "Bearer")
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
		isNotExp, err := cfg.db.RJWTNotExp(rToken.Raw)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("couldn't verify RJWT expiration status: %s", err.Error()))
			return
		}

		if !isNotExp {
			respondWithError(w, http.StatusUnauthorized, "RJWT is expired")
			return
		}

		// creates AJWT
		userId, err := strconv.Atoi(claims.Subject)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("couldn't get user: %s", err.Error()))
			return
		}

		secsInHour := 3600
		aToken, err := auth.CreateAccessToken(userId, cfg.jwtSecret, int64(secsInHour))
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("couldn't create AJWT: %s", err.Error()))
			return
		}

		// responds with AJWT
		respondWithJSON(w, http.StatusOK, aToken)

	} else {
		respondWithError(w, http.StatusUnauthorized, "invalid RJWT")
	}
}
