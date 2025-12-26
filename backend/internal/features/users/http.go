package users

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jobping/backend/internal/middleware"
	"github.com/jobping/backend/internal/shared"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteError(w, shared.ErrBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		shared.WriteError(w, shared.NewAPIError(http.StatusBadRequest, "username and password are required"))
		return
	}

	resp, err := h.service.Register(r.Context(), req)
	if err != nil {
		if apiErr, ok := err.(*shared.APIError); ok {
			shared.WriteError(w, apiErr)
			return
		}
		shared.WriteError(w, shared.ErrInternalError)
		return
	}

	shared.WriteJSON(w, http.StatusCreated, resp)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteError(w, shared.ErrBadRequest)
		return
	}

	resp, err := h.service.Login(r.Context(), req)
	if err != nil {
		if apiErr, ok := err.(*shared.APIError); ok {
			shared.WriteError(w, apiErr)
			return
		}
		shared.WriteError(w, shared.ErrInternalError)
		return
	}

	shared.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		shared.WriteError(w, shared.ErrUnauthorized)
		return
	}

	resp, err := h.service.GetPreferences(r.Context(), userID)
	if err != nil {
		shared.WriteError(w, shared.ErrInternalError)
		return
	}

	shared.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) CreatePreference(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		shared.WriteError(w, shared.ErrUnauthorized)
		return
	}

	var req CreatePreferenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteError(w, shared.ErrBadRequest)
		return
	}

	if req.Key == "" || req.Value == "" {
		shared.WriteError(w, shared.NewAPIError(http.StatusBadRequest, "key and value are required"))
		return
	}

	resp, err := h.service.CreatePreference(r.Context(), userID, req)
	if err != nil {
		if apiErr, ok := err.(*shared.APIError); ok {
			shared.WriteError(w, apiErr)
			return
		}
		shared.WriteError(w, shared.ErrInternalError)
		return
	}

	shared.WriteJSON(w, http.StatusCreated, resp)
}

func (h *Handler) UpdatePreference(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		shared.WriteError(w, shared.ErrUnauthorized)
		return
	}

	prefID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		shared.WriteError(w, shared.ErrBadRequest)
		return
	}

	var req UpdatePreferenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteError(w, shared.ErrBadRequest)
		return
	}

	resp, err := h.service.UpdatePreference(r.Context(), userID, prefID, req)
	if err != nil {
		if apiErr, ok := err.(*shared.APIError); ok {
			shared.WriteError(w, apiErr)
			return
		}
		shared.WriteError(w, shared.ErrInternalError)
		return
	}

	shared.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) DeletePreference(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		shared.WriteError(w, shared.ErrUnauthorized)
		return
	}

	prefID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		shared.WriteError(w, shared.ErrBadRequest)
		return
	}

	if err := h.service.DeletePreference(r.Context(), userID, prefID); err != nil {
		if apiErr, ok := err.(*shared.APIError); ok {
			shared.WriteError(w, apiErr)
			return
		}
		shared.WriteError(w, shared.ErrInternalError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

