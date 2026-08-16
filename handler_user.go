package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Chirpy/internal/auth"
	"github.com/Chirpy/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
	Token        string    `json:"token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
}

func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}

	if err := decoder.Decode(&params); err != nil {
		responseWithError(w, http.StatusInternalServerError, "Couldn't decode parametes", err)
		return
	}

	if params.Email == "" {
		responseWithError(w, http.StatusBadRequest, "Email in required", nil)
		return
	}

	if params.Password == "" {
		responseWithError(w, http.StatusBadRequest, "Password is required", nil)
		return
	}

	hashedPass, err := auth.HashPassword(params.Password)

	if err != nil {
		responseWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
	}

	user, err := cfg.db.CreateUser(req.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPass,
	})

	if err != nil {
		responseWithError(w, http.StatusInternalServerError, "Couldn't create user", err)
		return
	}

	newUser := User{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	}

	responseWithJSON(w, http.StatusCreated, newUser)
}

func (cfg *apiConfig) handlerUsersLogin(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}

	if err := decoder.Decode(&params); err != nil {
		responseWithError(w, http.StatusInternalServerError, "Couldn't decode parametes", err)
		return
	}

	if params.Email == "" {
		responseWithError(w, http.StatusBadRequest, "Email in required", nil)
		return
	}

	if params.Password == "" {
		responseWithError(w, http.StatusBadRequest, "Password is required", nil)
		return
	}

	user, err := cfg.db.GetUserByEmail(req.Context(), params.Email)

	if err != nil {
		responseWithError(w, http.StatusInternalServerError, "Failed to fetch user", err)
		return
	}

	success, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)

	if err != nil {
		responseWithError(w, http.StatusInternalServerError, "Failed to check password", err)
		return
	}

	if !success {
		responseWithError(w, http.StatusUnauthorized, "Incorrect email or password", nil)
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Hour)

	if err != nil || token == "" {
		responseWithError(w, http.StatusInternalServerError, "Failed to create token", err)
		return
	}

	makeRefreshToken := auth.MakeRefreshToken()

	refreshToken, err := cfg.db.CreateRefreshToken(req.Context(), database.CreateRefreshTokenParams{
		Token:     makeRefreshToken,
		ExpiresAt: time.Now().UTC().Add(time.Hour * 24 * 60),
		UserID:    user.ID,
	})

	if err != nil {
		responseWithError(w, http.StatusInternalServerError, "Failed to save refresh token", err)
		return
	}

	responseWithJSON(w, http.StatusOK, User{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		IsChirpyRed:  user.IsChirpyRed,
		Token:        token,
		RefreshToken: refreshToken.Token,
	})

}

func (cfg *apiConfig) handlerUsersUpdate(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

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

	decoder := json.NewDecoder(req.Body)
	params := parameters{}

	if err := decoder.Decode(&params); err != nil {
		responseWithError(w, http.StatusInternalServerError, "Couldn't decode parametes", err)
		return
	}

	if params.Email == "" {
		responseWithError(w, http.StatusBadRequest, "Email in required", nil)
		return
	}

	if params.Password == "" {
		responseWithError(w, http.StatusBadRequest, "Password is required", nil)
		return
	}

	hashedPass, err := auth.HashPassword(params.Password)

	if err != nil {
		responseWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
	}

	user, err := cfg.db.UpdateUserData(req.Context(), database.UpdateUserDataParams{
		Email:          params.Email,
		HashedPassword: hashedPass,
		ID:             userID,
	})

	if err != nil {
		responseWithError(w, http.StatusInternalServerError, "Couldn't update user", err)
		return
	}

	responseWithJSON(w, http.StatusOK, User{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	})

}
