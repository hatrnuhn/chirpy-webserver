package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/hatrnuhn/rssagg/internal/database"
	"golang.org/x/crypto/bcrypt"
)

func handleLogin(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	dat, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, 500, "couldn't read request")
		return
	}

	req := struct {
		Email    string
		Password string
	}{}

	err = json.Unmarshal(dat, &req)
	if err != nil {
		respondWithError(w, 400, "couldn't unmarshal request")
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

	var user database.User
	found := false

	for _, u := range users {
		if req.Email == u.Email {
			user = u
			found = true
			break
		}
	}

	if !found {
		respondWithError(w, 400, "couldn't not find user with such email")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		respondWithError(w, 401, "password doesn't not match!")
		return
	}

	respondWithJSON(w, 200, struct {
		ID    int    `json:"id"`
		Email string `json:"email"`
	}{
		ID:    user.ID,
		Email: user.Email,
	})
}
