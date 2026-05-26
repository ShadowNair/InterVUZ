package http

import (
	"net/http"
	"time"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/platform/httpjson"
	roomusecase "github.com/GIT_USER_ID/GIT_REPO_ID/internal/usecase/room"
)

type ListBookingsHandler struct {
	useCase *roomusecase.UseCase
}

func NewListBookingsHandler(useCase *roomusecase.UseCase) *ListBookingsHandler {
	return &ListBookingsHandler{useCase: useCase}
}

func (h *ListBookingsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	date, err := time.ParseInLocation("2006-01-02", r.URL.Query().Get("date"), time.Local)
	if err != nil {
		httpjson.WriteError(w, http.StatusBadRequest, "bad_request", "date must be YYYY-MM-DD")
		return
	}

	response, err := h.useCase.ListBookings(r.Context(), date)
	if err != nil {
		httpjson.WriteError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	httpjson.WriteJSON(w, http.StatusOK, response)
}
