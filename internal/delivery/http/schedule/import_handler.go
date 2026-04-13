package http

import (
	"encoding/json"
	"net/http"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/platform/httpjson"
	scheduleusecase "github.com/GIT_USER_ID/GIT_REPO_ID/internal/usecase/schedule"
)

type ImportHandler struct {
	useCase *scheduleusecase.UseCase
}

func NewImportHandler(useCase *scheduleusecase.UseCase) *ImportHandler {
	return &ImportHandler{useCase: useCase}
}

func (h *ImportHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var request domain.ScheduleImportRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httpjson.WriteError(w, http.StatusBadRequest, "bad_request", "invalid json body")
		return
	}

	response, err := h.useCase.Import(r.Context(), request)
	if err != nil {
		httpjson.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to import schedule")
		return
	}

	httpjson.WriteJSON(w, http.StatusAccepted, response)
}
