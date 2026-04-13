package http

import (
	"errors"
	"net/http"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/platform/httpjson"
	scheduleusecase "github.com/GIT_USER_ID/GIT_REPO_ID/internal/usecase/schedule"
)

type GroupScheduleHandler struct {
	useCase *scheduleusecase.UseCase
}

func NewGroupScheduleHandler(useCase *scheduleusecase.UseCase) *GroupScheduleHandler {
	return &GroupScheduleHandler{useCase: useCase}
}

func (h *GroupScheduleHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	groupID := r.PathValue("groupID")
	if groupID == "" {
		httpjson.WriteError(w, http.StatusBadRequest, "bad_request", "group id is required")
		return
	}

	response, err := h.useCase.GetGroupSchedule(r.Context(), groupID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			httpjson.WriteError(w, http.StatusNotFound, "not_found", "group schedule not found")
			return
		}

		httpjson.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to load group schedule")
		return
	}

	httpjson.WriteJSON(w, http.StatusOK, response)
}
