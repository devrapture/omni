package apperrors

import "errors"

var (
	// ErrUserNotFound is returned when a user is not found
	ErrUserNotFound = errors.New("user not found")
	// ErrInvalidToken is returned when a token is invalid
	ErrInvalidToken = errors.New("invalid token")
	// ErrSettingsNotFound is returned when a user settings are not found
	ErrSettingsNotFound = errors.New("settings not found")
	// ErrEmptyAPIKey is returned when a user does not provide an api key when mode is user key
	ErrEmptyAPIKey = errors.New("api key is required when mode is user key")
	// ErrNotSupportFile is returned when a file is not supported
	ErrNotSupportFile = errors.New("unsupported file type")
	// ErrEmptyCsvFile is returned when a csv file is empty
	ErrEmptyCsvFile = errors.New("csv file is empty")
	// ErrFileNameTooLong is returned when a file name is too long
	ErrFileNameTooLong = errors.New("File name is too long")
)
