package health

import (
	"net/http"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/platform/httpjson"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	httpjson.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
