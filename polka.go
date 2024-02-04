package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/hatrnuhn/rssagg/internal/database"
	"github.com/hatrnuhn/rssagg/internal/webhooks"
)

func (cfg *apiConfig) handlePostPolkaWebhooks(w http.ResponseWriter, r *http.Request) {
	// read and parse request json
	defer r.Body.Close()

	dat, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "couldn't read request")
		return
	}
	req := webhooks.PolkaReq{}
	err = json.Unmarshal(dat, &req)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "couldn't unmarshal request")
		return
	}

	// check if event is user.upgraded
	if req.Event != "user.upgraded" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		return
	}
	// check if user_id exists
	reqID := req.Data["user_id"]
	us, err := cfg.db.GetUsers()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't read database")
		return
	}

	var newU *database.User
	found := false
	for _, u := range us {
		if u.ID == reqID {
			found = true
			newU = &u
			break
		}
	}

	if !found {
		respondWithError(w, http.StatusNotFound, "user doesn't exist in db")
		return
	}

	// update user
	newU.IsChirpyRed = true
	_, err = cfg.db.UpdateUser(newU, false)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't update user")
		return
	}

	w.WriteHeader(http.StatusOK)
}
