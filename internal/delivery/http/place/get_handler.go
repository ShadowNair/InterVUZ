package http

import (
	"errors"
	"net/http"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/platform/httpjson"
	placeusecase "github.com/GIT_USER_ID/GIT_REPO_ID/internal/usecase/place"
)

type GetHandler struct {
	useCase *placeusecase.UseCase
}

func NewGetHandler(useCase *placeusecase.UseCase) *GetHandler {
	return &GetHandler{useCase: useCase}
}

func (h *GetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	placeID := r.PathValue("placeID")
	if placeID == "" {
		httpjson.WriteError(w, http.StatusBadRequest, "bad_request", "place id is required")
		return
	}

	place, err := h.useCase.GetByID(r.Context(), placeID)
	if err != nil {
		if errors.Is(err, placeusecase.ErrNotFound) {
			httpjson.WriteError(w, http.StatusNotFound, "not_found", "place not found")
			return
		}

		httpjson.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to load place")
		return
	}

	httpjson.WriteJSON(w, http.StatusOK, place)
}
