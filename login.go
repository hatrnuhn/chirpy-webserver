package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hatrnuhn/rssagg/internal/database"
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
		expiresInSecs = 86400
	}

	token, err := createToken(userID, cfg.jwtSecret, int64(expiresInSecs))
	if err != nil {
		respondWithError(w, 500, "couldn't create token")
		return
	}

	respondWithJSON(w, 200, struct {
		ID    int    `json:"id"`
		Email string `json:"email"`
		Token string `json:"token"`
	}{
		ID:    user.ID,
		Email: user.Email,
		Token: token,
	})
}

func createToken(userID int, secretKey string, expiresInSeconds int64) (string, error) {
	// Create the Claims
	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Duration(expiresInSeconds) * time.Second)),
		Subject:   strconv.Itoa(userID), // Convert userID to string
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return ss, nil
}
