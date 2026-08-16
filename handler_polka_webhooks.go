package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/Chirpy/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerPolkaWebhook(w http.ResponseWriter, req *http.Request) {
	type dataType struct {
		UserID string `json:"user_id"`
	}

	type parameters struct {
		Event string   `json:"event"`
		Data  dataType `json:"data"`
	}

	key, err := auth.GetAPIKey(req.Header)

	if err != nil {
		responseWithError(w, http.StatusUnauthorized, "Failed to verify access", err)
		return
	}

	if key != cfg.polkaKey {
		responseWithError(w, http.StatusUnauthorized, "Not allowed", nil)
		return
	}

	params := parameters{}

	if err := json.NewDecoder(req.Body).Decode(&params); err != nil {
		responseWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	_, err = cfg.db.UpgradeUserToRed(req.Context(), uuid.MustParse(params.Data.UserID))

	if err == sql.ErrNoRows {
		responseWithError(w, http.StatusNotFound, "User not found", nil)
		return
	} else if err != nil {
		responseWithError(w, http.StatusInternalServerError, "Couldn't upgrade user to red", err)
		return
	}

	responseWithJSON(w, http.StatusNoContent, nil)

}
