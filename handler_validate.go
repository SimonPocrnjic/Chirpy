package main

import (
	"encoding/json"
	"net/http"
	"regexp"
)

func handlerValidateChirp(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	type response struct {
		CleanedBody string `json:"cleaned_body"`
	}

	w.Header().Set("Content-Type", "application/json")

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

	responseWithJSON(w, http.StatusOK, response{
		CleanedBody: cleanedBody,
	})
}

func getCleanedBody(body string, badWords []string) string {
	cleaned := body

	for _, word := range badWords {
		re := regexp.MustCompile(`(?i)`+word)
		cleaned = re.ReplaceAllString(cleaned, "****")
	}

	return cleaned
}