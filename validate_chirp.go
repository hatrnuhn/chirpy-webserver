package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

func handleValidateChirp(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	type requestBody struct {
		Body string `json:"body"`
	}

	type responseBody struct {
		Msg string `json:"cleaned_body"`
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

	msg := strings.Split(req.Body, " ")

	profaneWords := []string{
		"kerfuffle",
		"sharbert",
		"fornax",
	}

	for i, word := range msg {
		w := strings.ToLower(word)
		for _, filWord := range profaneWords {
			if w == filWord {
				msg[i] = "****"
				break
			}
		}
	}

	respMsg := strings.Join(msg, " ")

	respondWithJSON(w, 200, responseBody{
		Msg: respMsg,
	})

}
