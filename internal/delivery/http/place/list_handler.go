package http

import (
	"net/http"
	"strconv"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/platform/httpjson"
	placeusecase "github.com/GIT_USER_ID/GIT_REPO_ID/internal/usecase/place"
)

type ListHandler struct {
	useCase *placeusecase.UseCase
}

func NewListHandler(useCase *placeusecase.UseCase) *ListHandler {
	return &ListHandler{useCase: useCase}
}

func (h *ListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	filter := domain.PlaceFilter{
		Type:     r.URL.Query().Get("type"),
		Building: r.URL.Query().Get("building"),
		Search:   r.URL.Query().Get("search"),
	}

	if floorRaw := r.URL.Query().Get("floor"); floorRaw != "" {
		floor, err := strconv.Atoi(floorRaw)
		if err != nil {
			httpjson.WriteError(w, http.StatusBadRequest, "bad_request", "floor must be an integer")
			return
		}
		filter.Floor = &floor
	}

	response, err := h.useCase.List(r.Context(), filter)
	if err != nil {
		httpjson.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to load places")
		return
	}

	httpjson.WriteJSON(w, http.StatusOK, response)
}
