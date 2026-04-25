package http

import (
	"errors"
	"net/http"
	"time"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/platform/httpjson"
	roomusecase "github.com/GIT_USER_ID/GIT_REPO_ID/internal/usecase/room"
)

type GetScheduleHandler struct {
	useCase *roomusecase.UseCase
}

func NewGetScheduleHandler(useCase *roomusecase.UseCase) *GetScheduleHandler {
	return &GetScheduleHandler{useCase: useCase}
}

func (h *GetScheduleHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("roomID")
	if roomID == "" {
		httpjson.WriteError(w, http.StatusBadRequest, "bad_request", "room id is required")
		return
	}

	dateRaw := r.URL.Query().Get("date")
	date, err := time.ParseInLocation("2006-01-02", dateRaw, time.Local)
	if err != nil {
		httpjson.WriteError(w, http.StatusBadRequest, "bad_request", "date must be YYYY-MM-DD")
		return
	}

	response, err := h.useCase.GetSchedule(r.Context(), roomID, date)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			httpjson.WriteError(w, http.StatusNotFound, "not_found", "room not found")
			return
		}

		httpjson.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to load room schedule")
		return
	}

	httpjson.WriteJSON(w, http.StatusOK, response)
}
