package http

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/platform/httpjson"
	roomusecase "github.com/GIT_USER_ID/GIT_REPO_ID/internal/usecase/room"
)

type ListAvailabilityHandler struct {
	useCase *roomusecase.UseCase
}

func NewListAvailabilityHandler(useCase *roomusecase.UseCase) *ListAvailabilityHandler {
	return &ListAvailabilityHandler{useCase: useCase}
}

func (h *ListAvailabilityHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	startsAt, err := time.Parse(time.RFC3339, r.URL.Query().Get("startsAt"))
	if err != nil {
		httpjson.WriteError(w, http.StatusBadRequest, "bad_request", "startsAt must be RFC3339 datetime")
		return
	}

	filter := domain.RoomAvailabilityFilter{
		StartsAt: startsAt,
		Building: r.URL.Query().Get("building"),
	}

	if endsAtRaw := r.URL.Query().Get("endsAt"); endsAtRaw != "" {
		endsAt, err := time.Parse(time.RFC3339, endsAtRaw)
		if err != nil {
			httpjson.WriteError(w, http.StatusBadRequest, "bad_request", "endsAt must be RFC3339 datetime")
			return
		}
		filter.EndsAt = endsAt
	}

	if capacityRaw := r.URL.Query().Get("capacity"); capacityRaw != "" {
		capacity, err := strconv.Atoi(capacityRaw)
		if err != nil {
			httpjson.WriteError(w, http.StatusBadRequest, "bad_request", "capacity must be an integer")
			return
		}
		filter.Capacity = &capacity
	}

	if equipmentRaw := r.URL.Query().Get("equipment"); equipmentRaw != "" {
		filter.Equipment = splitCSV(equipmentRaw)
	}

	response, err := h.useCase.ListAvailability(r.Context(), filter)
	if err != nil {
		httpjson.WriteError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	httpjson.WriteJSON(w, http.StatusOK, response)
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
