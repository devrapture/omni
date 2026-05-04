package apperrors

import "errors"

var (
	// ErrUserNotFound is returned when a user is not found
	ErrUserNotFound = errors.New("user not found")
	// ErrInvalidToken is returned when a token is invalid
	ErrInvalidToken = errors.New("invalid token")
	// ErrSettingsNotFound is returned when a user settings are not found
	ErrSettingsNotFound = errors.New("settings not found")
)
