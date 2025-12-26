package handler

import "github.com/google/uuid"

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

type UserResponse struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
}

type CreatePreferenceRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type UpdatePreferenceRequest struct {
	Value string `json:"value"`
}

type PreferenceResponse struct {
	ID    uuid.UUID `json:"id"`
	Key   string    `json:"key"`
	Value string    `json:"value"`
}

type PreferencesResponse struct {
	Preferences []PreferenceResponse `json:"preferences"`
}

