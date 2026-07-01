package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/platform/httpjson"
	roomusecase "github.com/GIT_USER_ID/GIT_REPO_ID/internal/usecase/room"
)

type CreateBookingHandler struct {
	useCase *roomusecase.UseCase
}

func NewCreateBookingHandler(useCase *roomusecase.UseCase) *CreateBookingHandler {
	return &CreateBookingHandler{useCase: useCase}
}

func (h *CreateBookingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("roomID")
	if roomID == "" {
		httpjson.WriteError(w, http.StatusBadRequest, "bad_request", "room id is required")
		return
	}

	defer r.Body.Close()
	var payload domain.CreateRoomBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpjson.WriteError(w, http.StatusBadRequest, "bad_request", "request body must be valid JSON")
		return
	}

	startsAt, err := time.Parse(time.RFC3339, payload.StartsAt)
	if err != nil {
		httpjson.WriteError(w, http.StatusBadRequest, "bad_request", "startsAt must be RFC3339 datetime")
		return
	}

	booking, err := h.useCase.CreateBooking(r.Context(), domain.RoomBookingRequest{
		RoomID:        roomID,
		StartsAt:      startsAt,
		BookerName:    payload.BookerName,
		BookerContact: payload.BookerContact,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			httpjson.WriteError(w, http.StatusNotFound, "not_found", "Комната не найдена")
		case errors.Is(err, domain.ErrConflict):
			httpjson.WriteError(w, http.StatusConflict, "conflict", "Комната занята в это время")
		default:
			httpjson.WriteError(w, http.StatusBadRequest, "bad_request", err.Error())
		}
		return
	}

	httpjson.WriteJSON(w, http.StatusCreated, booking)
}
