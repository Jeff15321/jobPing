package users

import (
	"net/http"

	"github.com/jobping/backend/internal/shared"
)

var (
	ErrUserNotFound       = shared.NewAPIError(http.StatusNotFound, "user not found")
	ErrUserAlreadyExists  = shared.NewAPIError(http.StatusConflict, "username already exists")
	ErrInvalidCredentials = shared.NewAPIError(http.StatusUnauthorized, "invalid credentials")
	ErrPreferenceNotFound = shared.NewAPIError(http.StatusNotFound, "preference not found")
	ErrPreferenceExists   = shared.NewAPIError(http.StatusConflict, "preference with this key already exists")
)

