package common

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
)

func WriteJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func WriteMethodNotAllowed(w http.ResponseWriter) {
	WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
}

func WriteDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		WriteError(w, http.StatusNotFound, "not_found", err.Error())
	default:
		WriteError(w, http.StatusBadRequest, "bad_request", err.Error())
	}
}

func WriteError(w http.ResponseWriter, statusCode int, code string, message string) {
	WriteJSON(w, statusCode, domain.ErrorResponse{
		Code:    code,
		Message: message,
	})
}
