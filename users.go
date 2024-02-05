package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hatrnuhn/rssagg/internal/auth"
	"github.com/hatrnuhn/rssagg/internal/database"
)

func (cfg *apiConfig) handlePostUsers(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	dat, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "couldn't read request")
		return
	}

	req := struct {
		ID          int    `json:"id"`
		Email       string `json:"email"`
		IsChirpyRed bool   `json:"is_chirpy_red"`
	}{}
	err = json.Unmarshal(dat, &req)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "couldn't unmarshal request")
		return
	}

	if len(req.Email) > 140 {
		respondWithError(w, http.StatusBadRequest, "email address is too long!")
		return
	}

	users, err := cfg.db.GetUsers()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't get users")
		return
	}

	for _, u := range users {
		if req.Email == u.Email {
			respondWithError(w, http.StatusBadRequest, "email is already registered")
			return
		}
	}

	newU, err := cfg.db.CreateUser(string(dat))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't create user")
		return
	}

	req.ID = newU.ID

	respondWithJSON(w, 201, req)
}

func (cfg *apiConfig) handlePutUsers(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	token, err := auth.ParseReq(r, cfg.jwtSecret, "Bearer")
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok {
		if claims.Issuer != "chirpy-access" {
			respondWithError(w, http.StatusUnauthorized, "invalid AJWT")
			return
		}

		if !ok {
			respondWithError(w, http.StatusUnauthorized, "access token is expired")
			return
		}

		dat, err := io.ReadAll(r.Body)
		if err != nil {
			respondWithError(w, 500, fmt.Sprintf("couldn't read request: %s", err.Error()))
			return
		}
		req := database.User{}
		err = json.Unmarshal(dat, &req)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, fmt.Sprintf("couldn't unmarshal request: %s", err.Error()))
			return
		}

		if len(req.Email) > 140 {
			respondWithError(w, http.StatusBadRequest, "email address is too long!")
			return
		}

		users, err := cfg.db.GetUsers()
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("couldn't get users: %s", err.Error()))
		}

		for _, u := range users {
			if u.Email == req.Email {
				respondWithError(w, http.StatusBadRequest, "email already exists")
				return
			}
		}

		id, err := strconv.Atoi(claims.Subject)
		if err != nil {
			respondWithJSON(w, http.StatusBadRequest, "couldn't get id")
			return
		}

		req.ID = id

		resp, err := cfg.db.UpdateUser(&req, true)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("couldn't update : %s", err.Error()))
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
		respondWithError(w, http.StatusUnauthorized, "invalid AJWT token")
		return
	}
}
