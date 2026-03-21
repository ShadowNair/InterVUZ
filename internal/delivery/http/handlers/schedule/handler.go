package schedule

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

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
	// Исправленный путь (без ведущего слеша, если mux настроен так)
	if r.URL.Path != "/admin/schedule/import" && r.URL.Path != "admin/schedule/import" {
		http.NotFound(w, r)
		return
	}
	
	if r.Method != http.MethodPost {
		common.WriteMethodNotAllowed(w)
		return
	}

	// Опционально: добавляем таймаут на всю операцию импорта
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	response, err := h.useCase.Import(ctx, domain.ScheduleImportRequest{})
	if err != nil {
		// Map errors to appropriate HTTP status
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			common.WriteError(w, http.StatusRequestTimeout, "timeout", "import operation timed out")
			return
		}
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
