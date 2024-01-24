package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hatrnuhn/rssagg/internal/database"
)

func (cfg *apiConfig) handlePostUsers(w http.ResponseWriter, r *http.Request) {
	dat, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		respondWithError(w, 500, "couldn't read request")
		return
	}

	req := struct {
		ID    int    `json:"id"`
		Email string `json:"email"`
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

	users, err := cfg.db.GetUsers()
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

	newU, err := cfg.db.CreateUser(string(dat))
	if err != nil {
		respondWithError(w, 500, "couldn't create user")
		return
	}

	req.ID = newU.ID

	respondWithJSON(w, 201, req)
}

func (cfg *apiConfig) handlePutUsers(w http.ResponseWriter, r *http.Request) {
	authHead := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHead, "Bearer ") {
		respondWithError(w, 401, "invalid auth header")
		return
	}

	tokenString := strings.TrimPrefix(authHead, "Bearer ")

	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(cfg.jwtSecret), nil
	})

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok {
		if time.Now().Unix() > claims.ExpiresAt.Unix() {
			respondWithError(w, http.StatusUnauthorized, "token is expired")
			return
		}

		dat, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			respondWithError(w, 500, "couldn't read request")
			return
		}

		req := database.User{}
		err = json.Unmarshal(dat, &req)
		if err != nil {
			respondWithError(w, 500, "couldn't unmarshal request")
			return
		}

		if len(req.Email) > 140 {
			respondWithError(w, 400, "email address is too long!")
			return
		}

		id, err := strconv.Atoi(claims.Subject)
		if err != nil {
			respondWithJSON(w, 500, "couldn't get id")
			return
		}

		req.ID = id

		resp, err := cfg.db.UpdateUser(&req)
		if err != nil {
			respondWithError(w, 401, "couldn't update user")
			return
		}
		respondWithJSON(w, http.StatusOK, struct {
			ID    int    `json:"id"`
			Email string `json:"email"`
		}{
			ID:    resp.ID,
			Email: resp.Email,
		})
		return
	} else {
		respondWithError(w, http.StatusUnauthorized, "invalid JTW token")
		return
	}
}
