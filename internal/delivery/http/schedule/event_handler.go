package http

import (
	"errors"
	"net/http"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/platform/httpjson"
	scheduleusecase "github.com/GIT_USER_ID/GIT_REPO_ID/internal/usecase/schedule"
)

type EventHandler struct {
	useCase *scheduleusecase.UseCase
}

func NewEventHandler(useCase *scheduleusecase.UseCase) *EventHandler {
	return &EventHandler{useCase: useCase}
}

func (h *EventHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	eventID := r.PathValue("eventID")
	if eventID == "" {
		httpjson.WriteError(w, http.StatusBadRequest, "bad_request", "event id is required")
		return
	}

	response, err := h.useCase.GetEvent(r.Context(), eventID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			httpjson.WriteError(w, http.StatusNotFound, "not_found", "event not found")
			return
		}

		httpjson.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to load event")
		return
	}

	httpjson.WriteJSON(w, http.StatusOK, response)
}
