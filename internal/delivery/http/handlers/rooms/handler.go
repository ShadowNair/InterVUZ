package rooms

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/handlers/common"
	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/usecase"
)

type Handler struct {
	useCase *usecase.RoomUseCase
}

func NewHandler(useCase *usecase.RoomUseCase) *Handler {
	return &Handler{useCase: useCase}
}

func (h *Handler) ListAvailability(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/rooms/availability" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		common.WriteMethodNotAllowed(w)
		return
	}

	startsAt, err := time.Parse(time.RFC3339, r.URL.Query().Get("startsAt"))
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, "bad_request", "startsAt must be RFC3339 datetime")
		return
	}

	endsAt, err := time.Parse(time.RFC3339, r.URL.Query().Get("endsAt"))
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, "bad_request", "endsAt must be RFC3339 datetime")
		return
	}

	filter := domain.RoomAvailabilityFilter{
		StartsAt: startsAt,
		EndsAt:   endsAt,
		Building: r.URL.Query().Get("building"),
	}

	if capacityRaw := r.URL.Query().Get("capacity"); capacityRaw != "" {
		capacity, err := strconv.Atoi(capacityRaw)
		if err != nil {
			common.WriteError(w, http.StatusBadRequest, "bad_request", "capacity must be an integer")
			return
		}
		filter.Capacity = &capacity
	}

	if equipmentRaw := r.URL.Query().Get("equipment"); equipmentRaw != "" {
		filter.Equipment = splitCSV(equipmentRaw)
	}

	response, err := h.useCase.ListAvailability(r.Context(), filter)
	if err != nil {
		common.WriteDomainError(w, err)
		return
	}

	common.WriteJSON(w, http.StatusOK, response)
}

func splitCSV(value string) []string {
	rawItems := strings.Split(value, ",")
	items := make([]string, 0, len(rawItems))
	for _, item := range rawItems {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			items = append(items, trimmed)
		}
	}
	return items
}
