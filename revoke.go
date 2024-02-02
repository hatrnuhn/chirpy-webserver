package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hatrnuhn/rssagg/internal/auth"
)

func (cfg *apiConfig) handlePostRevoke(w http.ResponseWriter, r *http.Request) {
	// reads rToken from Header
	rToken, err := auth.ParseReq(r, cfg.jwtSecret)

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

		// updates tokens on db
		_, err = cfg.db.WriteRefreshToken(rToken.Raw, time.Now().Unix())
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
