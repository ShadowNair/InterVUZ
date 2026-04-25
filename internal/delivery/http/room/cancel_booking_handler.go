package http

import (
	"errors"
	"net/http"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/platform/httpjson"
	roomusecase "github.com/GIT_USER_ID/GIT_REPO_ID/internal/usecase/room"
)

type CancelBookingHandler struct {
	useCase *roomusecase.UseCase
}

func NewCancelBookingHandler(useCase *roomusecase.UseCase) *CancelBookingHandler {
	return &CancelBookingHandler{useCase: useCase}
}

func (h *CancelBookingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	bookingID := r.PathValue("bookingID")
	if bookingID == "" {
		httpjson.WriteError(w, http.StatusBadRequest, "bad_request", "booking id is required")
		return
	}

	if err := h.useCase.CancelBooking(r.Context(), bookingID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			httpjson.WriteError(w, http.StatusNotFound, "not_found", "booking not found")
			return
		}

		httpjson.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to cancel booking")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
