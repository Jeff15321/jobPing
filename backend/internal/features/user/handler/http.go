package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jobping/backend/internal/features/user/service"
	"github.com/jobping/backend/internal/features/user/usererr"
)

type UserHandler struct {
	service *service.UserService
	auth    *AuthMiddleware
}

func NewUserHandler(svc *service.UserService, auth *AuthMiddleware) *UserHandler {
	return &UserHandler{
		service: svc,
		auth:    auth,
	}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "username and password are required")
		return
	}

	userEntity, err := h.service.Register(r.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, usererr.ErrUserAlreadyExists) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	token, err := h.auth.GenerateToken(userEntity.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, AuthResponse{
		Token: token,
		User:  UserResponse{ID: userEntity.ID, Username: userEntity.Username},
	})
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userEntity, err := h.service.Authenticate(r.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, usererr.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	token, err := h.auth.GenerateToken(userEntity.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, AuthResponse{
		Token: token,
		User:  UserResponse{ID: userEntity.ID, Username: userEntity.Username},
	})
}

func (h *UserHandler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	prefs, err := h.service.GetPreferences(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	response := PreferencesResponse{Preferences: make([]PreferenceResponse, len(prefs))}
	for i, p := range prefs {
		response.Preferences[i] = PreferenceResponse{ID: p.ID, Key: p.Key, Value: p.Value}
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *UserHandler) CreatePreference(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreatePreferenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Key == "" || req.Value == "" {
		writeError(w, http.StatusBadRequest, "key and value are required")
		return
	}

	pref, err := h.service.CreatePreference(r.Context(), userID, req.Key, req.Value)
	if err != nil {
		if errors.Is(err, usererr.ErrPreferenceExists) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, PreferenceResponse{ID: pref.ID, Key: pref.Key, Value: pref.Value})
}

func (h *UserHandler) UpdatePreference(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	prefID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid preference id")
		return
	}

	var req UpdatePreferenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	pref, err := h.service.UpdatePreference(r.Context(), userID, prefID, req.Value)
	if err != nil {
		if errors.Is(err, usererr.ErrPreferenceNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, PreferenceResponse{ID: pref.ID, Key: pref.Key, Value: pref.Value})
}

func (h *UserHandler) DeletePreference(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	prefID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid preference id")
		return
	}

	if err := h.service.DeletePreference(r.Context(), userID, prefID); err != nil {
		if errors.Is(err, usererr.ErrPreferenceNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.service.GetUserByID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if user == nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	writeJSON(w, http.StatusOK, ProfileResponse{
		ID:              user.ID,
		Username:        user.Username,
		AIPrompt:        user.AIPrompt,
		DiscordWebhook:  user.DiscordWebhook,
		NotifyThreshold: user.NotifyThreshold,
	})
}

func (h *UserHandler) UpdatePrompt(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req UpdatePromptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.service.UpdateAIPrompt(r.Context(), userID, req.Prompt); err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "prompt updated"})
}

func (h *UserHandler) UpdateDiscord(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req UpdateDiscordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.service.UpdateDiscordWebhook(r.Context(), userID, req.WebhookURL); err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "discord webhook updated"})
}

func (h *UserHandler) UpdateThreshold(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req UpdateThresholdRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Threshold < 0 || req.Threshold > 100 {
		writeError(w, http.StatusBadRequest, "threshold must be between 0 and 100")
		return
	}

	if err := h.service.UpdateNotifyThreshold(r.Context(), userID, req.Threshold); err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "threshold updated"})
}

func (h *UserHandler) GetMatches(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	matches, err := h.service.GetUserMatches(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	response := UserMatchesResponse{Matches: make([]UserJobMatchResponse, len(matches))}
	for i, m := range matches {
		response.Matches[i] = UserJobMatchResponse{
			ID:        m.ID,
			JobID:     m.JobID,
			Score:     m.Score,
			Analysis:  m.Analysis,
			Notified:  m.Notified,
			CreatedAt: m.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	writeJSON(w, http.StatusOK, response)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{"code": status, "message": message})
}
