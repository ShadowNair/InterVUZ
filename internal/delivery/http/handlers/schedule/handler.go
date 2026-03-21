package schedule

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/handlers/common"
	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/usecase"
)

type Handler struct {
	useCase *usecase.ScheduleUseCase
}

func NewHandler(useCase *usecase.ScheduleUseCase) *Handler {
	return &Handler{useCase: useCase}
}

func (h *Handler) Import(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/schedule/import" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		common.WriteMethodNotAllowed(w)
		return
	}

	var request domain.ScheduleImportRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		common.WriteError(w, http.StatusBadRequest, "bad_request", "invalid json body")
		return
	}

	response, err := h.useCase.Import(r.Context(), request)
	if err != nil {
		common.WriteDomainError(w, err)
		return
	}

	common.WriteJSON(w, http.StatusAccepted, response)
}

func (h *Handler) GetGroups(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/users/schedule/groups" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		common.WriteMethodNotAllowed(w)
		return
	}

	response, err := h.useCase.GetGroups(r.Context())
	if err != nil {
		common.WriteDomainError(w, err)
		return
	}

	common.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) GetGroupSchedule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		common.WriteMethodNotAllowed(w)
		return
	}

	groupID := strings.TrimPrefix(r.URL.Path, "/users/schedule/")
	if groupID == "" || strings.Contains(groupID, "/") {
		http.NotFound(w, r)
		return
	}

	response, err := h.useCase.GetGroupSchedule(r.Context(), groupID)
	if err != nil {
		common.WriteDomainError(w, err)
		return
	}

	common.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) GetEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		common.WriteMethodNotAllowed(w)
		return
	}

	eventID := strings.TrimPrefix(r.URL.Path, "/users/schedule/events/")
	if eventID == "" || strings.Contains(eventID, "/") {
		http.NotFound(w, r)
		return
	}

	response, err := h.useCase.GetEvent(r.Context(), eventID)
	if err != nil {
		common.WriteDomainError(w, err)
		return
	}

	common.WriteJSON(w, http.StatusOK, response)
}
