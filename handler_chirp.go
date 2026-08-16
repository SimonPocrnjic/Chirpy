package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"regexp"
	"sort"
	"time"

	"github.com/Chirpy/internal/auth"
	"github.com/Chirpy/internal/database"
	"github.com/google/uuid"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerChirpsCreate(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	type response struct {
		CleanedBody string `json:"cleaned_body"`
	}

	bearerToken, err := auth.GetBearerToken(req.Header)

	if err != nil {
		responseWithError(w, http.StatusInternalServerError, "Couldn't retrive token", err)
		return
	}

	userID, err := auth.ValidateJWT(bearerToken, cfg.jwtSecret)

	if err != nil {
		responseWithError(w, http.StatusUnauthorized, "Failed to validate token", err)
		return
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		responseWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	if len(params.Body) > 140 {
		responseWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	cleanedBody := getCleanedBody(params.Body, []string{"kerfuffle", "sharbert", "fornax"})

	chirp, err := cfg.db.CreateChirp(req.Context(), database.CreateChirpParams{
		Body:   cleanedBody,
		UserID: userID,
	})

	if err != nil {
		responseWithError(w, http.StatusInternalServerError, "Couldn't create chirp", err)
		return
	}

	responseWithJSON(w, http.StatusCreated, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}

func (cfg *apiConfig) handlerChirpsGetAll(w http.ResponseWriter, req *http.Request) {

	authorID, err := authorIDFromRequest(req)
	if err != nil {
		responseWithError(w, http.StatusBadRequest, "Invalid author ID", err)
		return
	}

	sortDir := req.URL.Query().Get("sort")

	var dbChirps []database.Chirp

	if authorID != uuid.Nil {
		dbChirps, err = cfg.db.GetChirpsByUser(req.Context(), authorID)
	} else {
		dbChirps, err = cfg.db.GetChirps(req.Context())
	}

	if err != nil {
		responseWithError(w, http.StatusInternalServerError, "Couldn't fetch chirps", err)
		return
	}

	if sortDir != "" && sortDir == "desc" {
		sort.Slice(dbChirps, func(i, j int) bool {
			return dbChirps[i].CreatedAt.UTC().Compare(dbChirps[j].CreatedAt) == 1
		})
	}

	chirps := []Chirp{}

	for _, c := range dbChirps {
		chirps = append(chirps, Chirp{
			ID:        c.ID,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
			Body:      c.Body,
			UserID:    c.UserID,
		})
	}

	responseWithJSON(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) handlerChirpsGet(w http.ResponseWriter, req *http.Request) {

	chirpID := uuid.MustParse(req.PathValue("chirpID"))

	chirp, err := cfg.db.GetChirp(req.Context(), chirpID)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	responseWithJSON(w, http.StatusOK, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}

func (cfg *apiConfig) handlerChirpsDelete(w http.ResponseWriter, req *http.Request) {
	bearerToken, err := auth.GetBearerToken(req.Header)

	if err != nil {
		responseWithError(w, http.StatusUnauthorized, "Couldn't retrive token", err)
		return
	}

	userID, err := auth.ValidateJWT(bearerToken, cfg.jwtSecret)

	if err != nil {
		responseWithError(w, http.StatusUnauthorized, "Failed to validate token", err)
		return
	}

	chirpID := uuid.MustParse(req.PathValue("chirpID"))

	chirp, err := cfg.db.GetChirp(req.Context(), chirpID)

	if err == sql.ErrNoRows {
		responseWithError(w, http.StatusNotFound, "Couldn't find chirp", err)
		return
	} else if err != nil {
		responseWithError(w, http.StatusInternalServerError, "Something went wrong", err)
	}

	if chirp.UserID != userID {
		responseWithError(w, http.StatusForbidden, "User not allowed to delete record", nil)
		return
	}

	if err := cfg.db.DeleteChirp(req.Context(), chirp.ID); err != nil {
		responseWithError(w, http.StatusInternalServerError, "Failed to delete record", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)

}

func getCleanedBody(body string, badWords []string) string {
	cleaned := body

	for _, word := range badWords {
		re := regexp.MustCompile(`(?i)` + word)
		cleaned = re.ReplaceAllString(cleaned, "****")
	}

	return cleaned
}

func authorIDFromRequest(req *http.Request) (uuid.UUID, error) {
	authorIDString := req.URL.Query().Get("author_id")
	if authorIDString == "" {
		return uuid.Nil, nil
	}

	authorID, err := uuid.Parse(authorIDString)
	if err != nil {
		return uuid.Nil, err
	}
	return authorID, nil
}
