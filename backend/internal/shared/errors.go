package shared

import (
	"encoding/json"
	"net/http"
)

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	return e.Message
}

func NewAPIError(code int, message string) *APIError {
	return &APIError{Code: code, Message: message}
}

var (
	ErrNotFound      = NewAPIError(http.StatusNotFound, "resource not found")
	ErrUnauthorized  = NewAPIError(http.StatusUnauthorized, "unauthorized")
	ErrBadRequest    = NewAPIError(http.StatusBadRequest, "bad request")
	ErrInternalError = NewAPIError(http.StatusInternalServerError, "internal server error")
)

func WriteError(w http.ResponseWriter, err *APIError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Code)
	json.NewEncoder(w).Encode(err)
}

func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
