package apperrors

import (
	"errors"
	"strings"
)

const (
	CodeGeminiKeyMissing       = "GEMINI_KEY_MISSING"
	CodeGeminiKeyInvalid       = "GEMINI_KEY_INVALID"
	CodeGeminiKeyQuotaExceeded = "GEMINI_KEY_QUOTA_EXCEEDED"
	CodeGeminiKeyError         = "GEMINI_KEY_ERROR"
)

func GeminiErrorCode(err error) string {
	switch {
	case errors.Is(err, ErrMissingUserGeminiKey):
		return CodeGeminiKeyMissing
	case errors.Is(err, ErrInvalidGeminiKey):
		return CodeGeminiKeyInvalid
	case errors.Is(err, ErrGeminiQuotaExceeded):
		return CodeGeminiKeyQuotaExceeded
	case errors.Is(err, ErrGeminiKeyRejected):
		return CodeGeminiKeyError
	default:
		return ""
	}
}

func GeminiErrorMessage(code string) string {
	switch code {
	case CodeGeminiKeyMissing:
		return "Please add your Gemini API key in settings."
	case CodeGeminiKeyInvalid:
		return "Your Gemini API key is invalid. Please update it in settings."
	case CodeGeminiKeyQuotaExceeded:
		return "Your Gemini API key has exhausted its quota or rate limit."
	case CodeGeminiKeyError:
		return "Gemini rejected your API key. Please check your key and billing settings."
	default:
		return ""
	}
}

func UserFacingGeminiError(err error) (string, string, bool) {
	code := GeminiErrorCode(err)
	if code == "" {
		return "", "", false
	}
	return code, GeminiErrorMessage(code), true
}

func UserFacingStoredGeminiError(raw string) (string, string, bool) {
	value := strings.ToLower(raw)

	switch {
	case strings.Contains(value, strings.ToLower(CodeGeminiKeyMissing)),
		strings.Contains(value, ErrMissingUserGeminiKey.Error()):
		return CodeGeminiKeyMissing, GeminiErrorMessage(CodeGeminiKeyMissing), true
	case strings.Contains(value, strings.ToLower(CodeGeminiKeyInvalid)),
		strings.Contains(value, ErrInvalidGeminiKey.Error()),
		strings.Contains(value, "api_key_invalid"),
		strings.Contains(value, "api key not valid"),
		strings.Contains(value, "invalid api key"):
		return CodeGeminiKeyInvalid, GeminiErrorMessage(CodeGeminiKeyInvalid), true
	case strings.Contains(value, strings.ToLower(CodeGeminiKeyQuotaExceeded)),
		strings.Contains(value, ErrGeminiQuotaExceeded.Error()),
		strings.Contains(value, "resource_exhausted"),
		strings.Contains(value, "quota"),
		strings.Contains(value, "rate limit"):
		return CodeGeminiKeyQuotaExceeded, GeminiErrorMessage(CodeGeminiKeyQuotaExceeded), true
	case strings.Contains(value, strings.ToLower(CodeGeminiKeyError)),
		strings.Contains(value, ErrGeminiKeyRejected.Error()):
		return CodeGeminiKeyError, GeminiErrorMessage(CodeGeminiKeyError), true
	default:
		return "", "", false
	}
}
