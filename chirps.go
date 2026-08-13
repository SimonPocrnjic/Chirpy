package main

import (
	"encoding/json"
	"net/http"
	"regexp"
	"time"

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
		Body   string    `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}

	type response struct {
		CleanedBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
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
		UserID: params.UserId,
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

	dbChirps, err := cfg.db.GetChirps(req.Context())

	if err != nil {
		responseWithError(w, http.StatusInternalServerError, "Couldn't fetch chirps", err)
		return
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

func getCleanedBody(body string, badWords []string) string {
	cleaned := body

	for _, word := range badWords {
		re := regexp.MustCompile(`(?i)` + word)
		cleaned = re.ReplaceAllString(cleaned, "****")
	}

	return cleaned
}
