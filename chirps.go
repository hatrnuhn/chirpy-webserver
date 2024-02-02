package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/hatrnuhn/rssagg/internal/database"
)

// TODO: create a handleFunc that handles GET to /api/chirps
// TODO: only allows users create chirps as long as they're authenticated
// TODO: how to store authentication??
// TODO: associates new chirps to users who created them

// accepts and store a chirp POST and responds with a newly stored chirp
func (cfg *apiConfig) handlePostChirps(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	dat, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, 500, "couldn't read request")
		return
	}

	// authentication check -----
	// retrieve atoken from db

	// check atoken validity
	// use data from atoken to get userid and to associate new chirp

	req := database.Chirp{}
	err = json.Unmarshal(dat, &req)
	if err != nil {
		respondWithError(w, 500, "couldn't unmarshal request")
		return
	}

	if len(req.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long!")
		return
	}

	newC, err := cfg.db.CreateChirp(string(dat))
	if err != nil {
		respondWithError(w, 500, "couldn't create chirp")
		return
	}

	respondWithJSON(w, 201, newC)
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
