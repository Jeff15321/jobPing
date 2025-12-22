package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/yourusername/ai-job-scanner/internal/database"
)

type UserHandler struct {
	db *database.DB
}

func NewUserHandler(db *database.DB) *UserHandler {
	return &UserHandler{db: db}
}

type CreateUserRequest struct {
	Email       string `json:"email"`
	Preferences string `json:"preferences"`
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Implement user creation
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User created successfully",
		"email":   req.Email,
	})
}

func (h *UserHandler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement preference update
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Preferences updated successfully",
	})
}
