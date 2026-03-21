package routes

import (
	"encoding/json"
	"net/http"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/handlers/common"
	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/usecase"
)

type Handler struct {
	useCase *usecase.RouteUseCase
}

func NewHandler(useCase *usecase.RouteUseCase) *Handler {
	return &Handler{useCase: useCase}
}

func (h *Handler) Build(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/routes" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		common.WriteMethodNotAllowed(w)
		return
	}

	var request domain.RouteRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		common.WriteError(w, http.StatusBadRequest, "bad_request", "invalid json body")
		return
	}

	response, err := h.useCase.Build(r.Context(), request)
	if err != nil {
		common.WriteDomainError(w, err)
		return
	}

	common.WriteJSON(w, http.StatusOK, response)
}
