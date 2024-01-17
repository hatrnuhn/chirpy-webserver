package main

import (
	"encoding/json"
	"io"
	"net/http"
)

func handleValidateChirp(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	type requestBody struct {
		Body string `json:"body"`
	}

	type responseBody struct {
		Valid string `json:"valid"`
	}

	dat, err := io.ReadAll(r.Body)

	if err != nil {

		respondWithError(w, 500, "couldn't read request")

		return

	}

	req := requestBody{}

	err = json.Unmarshal(dat, &req)

	if err != nil {

		respondWithError(w, 500, "couldn't unmarshal request")

		return

	}

	if len(req.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long!")

		return
	}

	respondWithJSON(w, 200, responseBody{
		Valid: "true",
	})

}
