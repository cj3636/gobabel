package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/cj3636/gobabel/internal/engine"
)

type ErrorResponse struct {
	Error APIError `json:"error"`
}
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

var errorMessages = map[string]string{
	"invalid_json":         "Invalid JSON.",
	"invalid_character":    "Text contains characters outside the configured alphabet.",
	"text_too_long":        "Text is too long.",
	"empty_text":           "Text must not be empty.",
	"invalid_range":        "Invalid range.",
	"invalid_address":      "Invalid address.",
	"unsupported_version":  "Unsupported book version.",
	"unsupported_alphabet": "Unsupported alphabet.",
	"unsupported_codec":    "Unsupported codec.",
	"crypto_open_failed":   "Could not decrypt address.",
}

func writeErr(w http.ResponseWriter, status int, code, msg string, details any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{APIError{code, msg, details}})
}

func writeAPIError(w http.ResponseWriter, status int, err error, fallbackCode, fallbackMsg string) {
	code, details := apiError(err, fallbackCode)
	msg := errorMessages[code]
	if msg == "" {
		msg = fallbackMsg
	}
	writeErr(w, status, code, msg, details)
}

func apiError(err error, fallbackCode string) (string, any) {
	var coded interface {
		ErrorCode() string
		ErrorDetails() any
	}
	if errors.As(err, &coded) {
		return coded.ErrorCode(), coded.ErrorDetails()
	}
	for _, item := range []struct {
		target error
		code   string
	}{
		{engine.ErrEmptyText, "empty_text"},
		{engine.ErrTextTooLong, "text_too_long"},
		{engine.ErrInvalidRange, "invalid_range"},
		{engine.ErrInvalidAddress, "invalid_address"},
		{engine.ErrUnsupportedVersion, "unsupported_version"},
		{engine.ErrUnsupportedAlphabet, "unsupported_alphabet"},
		{engine.ErrUnsupportedCodec, "unsupported_codec"},
		{engine.ErrCryptoOpenFailed, "crypto_open_failed"},
	} {
		if errors.Is(err, item.target) {
			return item.code, nil
		}
	}
	if fallbackCode != "" {
		return fallbackCode, nil
	}
	return "invalid_address", nil
}
