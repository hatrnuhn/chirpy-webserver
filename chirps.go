package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/golang-jwt/jwt/v5"
	"github.com/hatrnuhn/rssagg/internal/auth"
	"github.com/hatrnuhn/rssagg/internal/database"
)

// requires body and authorization header, authenticates, then accepts and store a chirp POST and responds with a newly stored chirp with its associated author UserID
func (cfg *apiConfig) handlePostChirps(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	aToken, err := auth.ParseReq(r, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't parse from header")
		return
	}

	if claims, ok := aToken.Claims.(*jwt.RegisteredClaims); ok {
		if claims.Issuer != "chirpy-access" {
			respondWithError(w, http.StatusUnauthorized, "invalid AJWT")
			return
		}
		if !ok {
			respondWithError(w, http.StatusUnauthorized, "AJWT expired")
			return
		}

		dat, err := io.ReadAll(r.Body)
		if err != nil {
			respondWithError(w, 500, "couldn't read request")
			return
		}

		req := database.Chirp{}
		err = json.Unmarshal(dat, &req)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "couldn't unmarshal request")
			return
		}

		if len(req.Body) > 140 {
			respondWithError(w, http.StatusBadRequest, "Chirp is too long!")
			return
		}

		uID, err := strconv.Atoi(claims.Subject)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "couldn't read ID off token")
			return
		}

		newC, err := cfg.db.CreateChirp(req, uID)
		if err != nil {
			respondWithError(w, 500, "couldn't create chirp")
			return
		}

		respondWithJSON(w, 201, newC)

	} else {
		respondWithError(w, http.StatusUnauthorized, "please authenticate with the associated user")
	}
}

// responds with all chirps stored in database
func (cfg *apiConfig) handleGetChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.db.GetChirps()
	if err != nil {
		respondWithError(w, 500, "couldn't get chirps")
		return
	}

	respondWithJSON(w, 200, chirps)
}

// handles /chirps/{chirpID} endpoints
func (cfg *apiConfig) handleChirpID(w http.ResponseWriter, r *http.Request) {
	param := chi.URLParam(r, "chirpID")
	id, err := strconv.Atoi(param)
	if err != nil {
		respondWithError(w, 400, err.Error())
		return
	}
	if id == 0 {
		respondWithError(w, 400, "chirp id starts at 1")
		return
	}

	cps, err := cfg.db.GetChirps()
	if err != nil {
		respondWithError(w, 500, "couldn't get chirps")
		return
	}

	if id > len(cps) {
		respondWithError(w, 404, fmt.Sprintf("chirp with id: %v is not found", id))
		return
	}

	respondWithJSON(w, 200, cps[id-1])
}
