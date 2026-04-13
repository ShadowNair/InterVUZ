package httpjson

import (
	"encoding/json"
	"net/http"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
)

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func WriteError(w http.ResponseWriter, status int, code string, message string) {
	WriteJSON(w, status, domain.ErrorResponse{
		Code:    code,
		Message: message,
	})
}
