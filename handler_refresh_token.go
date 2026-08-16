package main

import (
	"net/http"
	"time"

	"github.com/Chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRefreshToken(w http.ResponseWriter, req *http.Request) {
	refreshToken, err := auth.GetBearerToken(req.Header)

	type response struct {
		Token string `json:"token"`
	}

	if err != nil {
		responseWithError(w, http.StatusBadRequest, "Bearer not provided", err)
		return
	}

	user, err := cfg.db.GetUserFromRefreshToken(req.Context(), refreshToken)

	if err != nil {
		responseWithError(w, http.StatusUnauthorized, "Couldn't get user for refresh token", err)
		return
	}

	newAccessToken, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Hour)

	if err != nil {
		responseWithError(w, http.StatusUnauthorized, "Couldn't validate token", err)
		return
	}

	responseWithJSON(w, http.StatusOK, response{
		Token: newAccessToken,
	})

}

func (cfg *apiConfig) handlerRefreshTokenRevoke(w http.ResponseWriter, req *http.Request) {
	refreshToken, err := auth.GetBearerToken(req.Header)

	if err != nil {
		responseWithError(w, http.StatusBadRequest, "Bearer not provided", err)
		return
	}

	if _, err = cfg.db.RevokeRefreshToken(req.Context(), refreshToken); err != nil {
		responseWithError(w, http.StatusInternalServerError, "Failed to revoke refresh token", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)

}
