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
	// ErrEmptyDocxFile is returned when a docx file has no readable text
	ErrEmptyDocxFile = errors.New("docx file is empty")
	// ErrFileNameTooLong is returned when a file name is too long
	ErrFileNameTooLong = errors.New("File name is too long")
	// ErrNotSupportedPdfFile is returned when a pdf file is not supported
	ErrNotSupportedPdfFile = errors.New("unsupported pdf file")
	// ErrInvalidGeminiKey is returned when a gemini api key is invalid
	ErrInvalidGeminiKey = errors.New("invalid gemini api key")
	// ErrGeminiQuotaExceeded is returned when a gemini api key quota is exceeded
	ErrGeminiQuotaExceeded = errors.New("gemini api key quota exceeded")
	// ErrGeminiKeyRejected is returned when a gemini api key was rejected by the provider
	ErrGeminiKeyRejected = errors.New("gemini api key was rejected")
	// ErrMissingUserGeminiKey is returned when a user does not have a gemini api key
	ErrMissingUserGeminiKey = errors.New("user does not have a gemini api key")
	// ErrInvalidTelegramBotToken is returned when a telegram bot token is invalid
	ErrInvalidTelegramBotToken = errors.New("invalid telegram bot token")
	// ErrTelegramBotTokenNotProvided is returned when a telegram bot token is not provided
	ErrTelegramBotTokenNotProvided = errors.New("telegram bot token is not provided")
	// ErrInvalidTelegramBotFormat is returned when a telegram bot token is invalid
	ErrInvalidTelegramBotFormat = errors.New("Telegram bot tokens look like: 7123456789:AAFxxxxxxxxxxxxx — get yours from @BotFather")
)
