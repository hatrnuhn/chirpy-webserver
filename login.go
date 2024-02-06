package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hatrnuhn/chirpy-webserver/internal/auth"
	"github.com/hatrnuhn/chirpy-webserver/internal/database"
	"golang.org/x/crypto/bcrypt"
)

type reqBody struct {
	Password string `json:"password"`
	Email    string `json:"email"`
	Exp      int    `json:"expires_in_seconds"`
}

func (cfg *apiConfig) handlePostLogin(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	dat, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, 500, "couldn't read request")
		return
	}

	req := reqBody{}

	err = json.Unmarshal(dat, &req)
	if err != nil {
		respondWithError(w, 400, "couldn't unmarshal request")
		return
	}

	users, err := cfg.db.GetUsers()
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
		respondWithError(w, 404, "couldn't not find user with such email")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		respondWithError(w, 401, "password doesn't not match")
		return
	}

	// If valid, create JWT
	userID := user.ID        // this would come from your user validation logic
	expiresInSecs := req.Exp // also can be configurable depending on your security policy
	if expiresInSecs == 0 {
		expiresInSecs = 3600
	}

	secsInMonth := 24 * 3600 * 30

	aToken, err := auth.CreateAccessToken(userID, cfg.jwtSecret, int64(expiresInSecs))
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("couldn't create access token: %s", err.Error()))
		return
	}

	/*
		_, err = cfg.db.WriteAccessToken(aToken)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "couldn't write AJWT to db")
			return
		}
	*/

	rToken, err := auth.CreateRefreshToken(userID, cfg.jwtSecret, int64(secsInMonth))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("couldn't create refresh token: %s", err.Error()))
		return
	}

	_, err = cfg.db.WriteRefreshToken(rToken, 0)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("couldn't write refresh token to database: %s", err.Error()))
		return
	}

	respondWithJSON(w, 200, struct {
		ID          int    `json:"id"`
		Email       string `json:"email"`
		AToken      string `json:"access_token"`
		RToken      string `json:"refresh_token"`
		IsChirpyRed bool   `json:"is_chirpy_red"`
	}{
		ID:          user.ID,
		Email:       user.Email,
		AToken:      aToken,
		RToken:      rToken,
		IsChirpyRed: user.IsChirpyRed,
	})
}
