package health

import (
	"net/http"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/handlers/common"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		common.WriteMethodNotAllowed(w)
		return
	}

	common.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
