package http

import (
	"net/http"
	"strconv"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/platform/httpjson"
	newsusecase "github.com/GIT_USER_ID/GIT_REPO_ID/internal/usecase/news"
)

type ListHandler struct {
	useCase *newsusecase.UseCase
}

func NewListHandler(useCase *newsusecase.UseCase) *ListHandler {
	return &ListHandler{useCase: useCase}
}

func (h *ListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	limit := 20
	if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil {
			httpjson.WriteError(w, http.StatusBadRequest, "bad_request", "limit must be an integer")
			return
		}
		limit = parsed
	}

	response, err := h.useCase.List(r.Context(), limit)
	if err != nil {
		httpjson.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to load news")
		return
	}

	httpjson.WriteJSON(w, http.StatusOK, response)
}
