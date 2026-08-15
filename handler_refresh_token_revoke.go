package main

import (
	"net/http"

	"github.com/Chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRefreshTokenRevoke(w http.ResponseWriter, req *http.Request) {
	refreshToken, err := auth.GetBearerToken(req.Header)

	if err != nil {
		responseWithError(w, http.StatusBadRequest, "Bearer not provided", err)
		return
	}

	if err != nil {
		responseWithError(w, http.StatusInternalServerError, "Couldn't fetch refresh token", err)
		return
	}

	if _, err = cfg.db.RevokeRefreshToken(req.Context(), refreshToken); err != nil {
		responseWithError(w, http.StatusInternalServerError, "Failed to revoke refresh token", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)

}
