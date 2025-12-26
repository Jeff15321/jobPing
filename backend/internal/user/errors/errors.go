package errors

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("username already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrPreferenceNotFound = errors.New("preference not found")
	ErrPreferenceExists   = errors.New("preference with this key already exists")
)

