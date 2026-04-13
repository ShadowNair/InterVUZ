package http

import (
	"net/http"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/platform/httpjson"
	graphusecase "github.com/GIT_USER_ID/GIT_REPO_ID/internal/usecase/graph"
)

type GetHandler struct {
	useCase *graphusecase.UseCase
}

func NewGetHandler(useCase *graphusecase.UseCase) *GetHandler {
	return &GetHandler{useCase: useCase}
}

func (h *GetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	graph, err := h.useCase.Get(r.Context())
	if err != nil {
		httpjson.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to load graph")
		return
	}

	httpjson.WriteJSON(w, http.StatusOK, graph)
}
