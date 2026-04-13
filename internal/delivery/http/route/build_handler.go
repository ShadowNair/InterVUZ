package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/platform/httpjson"
	routeusecase "github.com/GIT_USER_ID/GIT_REPO_ID/internal/usecase/route"
)

type BuildHandler struct {
	useCase *routeusecase.UseCase
}

func NewBuildHandler(useCase *routeusecase.UseCase) *BuildHandler {
	return &BuildHandler{useCase: useCase}
}

func (h *BuildHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var request domain.RouteRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httpjson.WriteError(w, http.StatusBadRequest, "bad_request", "invalid json body")
		return
	}

	route, err := h.useCase.Build(r.Context(), request)
	if err != nil {
		switch {
		case errors.Is(err, routeusecase.ErrInvalidInput):
			httpjson.WriteError(w, http.StatusBadRequest, "bad_request", err.Error())
		case errors.Is(err, routeusecase.ErrStartPlaceNotFound),
			errors.Is(err, routeusecase.ErrDestinationPlaceNotFound):
			httpjson.WriteError(w, http.StatusNotFound, "not_found", err.Error())
		case errors.Is(err, routeusecase.ErrStartVertexNotFound),
			errors.Is(err, routeusecase.ErrDestinationVertexNotFound),
			errors.Is(err, routeusecase.ErrDestinationWithoutVertex),
			errors.Is(err, routeusecase.ErrRouteNotFound):
			httpjson.WriteError(w, http.StatusUnprocessableEntity, "route_unavailable", err.Error())
		default:
			httpjson.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to build route")
		}
		return
	}

	httpjson.WriteJSON(w, http.StatusOK, route)
}
