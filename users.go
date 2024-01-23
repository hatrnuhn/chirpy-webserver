package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/hatrnuhn/rssagg/internal/database"
)

func handlePostUsers(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	dat, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, 500, "couldn't read request")
		return
	}

	req := struct {
		ID    int
		Email string
	}{}
	err = json.Unmarshal(dat, &req)
	if err != nil {
		respondWithError(w, 500, "couldn't unmarshal request")
		return
	}

	if len(req.Email) > 140 {
		respondWithError(w, 400, "email address is too long!")
		return
	}

	path := "internal/database/database.json"
	db, err := database.NewDB(path)
	if err != nil {
		respondWithError(w, 500, "couldn't initialize database")
		return
	}

	users, err := db.GetUsers()
	if err != nil {
		respondWithError(w, 500, "couldn't get users")
		return
	}

	for _, u := range users {
		if req.Email == u.Email {
			respondWithError(w, 400, "email is already registered")
			return
		}
	}

	newU, err := db.CreateUser(string(dat))
	if err != nil {
		respondWithError(w, 500, "couldn't create user")
		return
	}

	req.ID = newU.ID

	respondWithJSON(w, 201, req)
}
