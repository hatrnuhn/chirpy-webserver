package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/hatrnuhn/rssagg/internal/database"
)

// accepts and store a chirp POST and responds with a newly stored chirp
func handlePostChirps(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	dat, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, 500, "couldn't read request")
		return
	}

	req := database.ChirpReq{}
	err = json.Unmarshal(dat, &req)
	if err != nil {
		respondWithError(w, 500, "couldn't unmarshal request")
		return
	}

	if len(req.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long!")
		return
	}

	path := "internal/database/database.json"
	db, err := database.NewDB(path)
	if err != nil {
		respondWithError(w, 500, "couldn't initialize db")
		return
	}

	newC, err := db.CreateChirp(string(dat))
	if err != nil {
		respondWithError(w, 500, "couldn't create chirp")
		return
	}

	respondWithJSON(w, 201, newC)
}

// responds with all chirps stored in database
func handleGetChirps(w http.ResponseWriter, r *http.Request) {
	path := "internal/database/database.json"
	db, err := database.NewDB(path)
	if err != nil {
		respondWithError(w, 500, "couldn't initialize db")
		return
	}

	chirps, err := db.GetChirps()
	if err != nil {
		respondWithError(w, 500, "couldn't get chirps")
		return
	}

	respondWithJSON(w, 200, chirps)
}
