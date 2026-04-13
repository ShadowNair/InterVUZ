package http

import (
	"net/http"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/platform/httpjson"
	scheduleusecase "github.com/GIT_USER_ID/GIT_REPO_ID/internal/usecase/schedule"
)

type GroupCatalogHandler struct {
	useCase *scheduleusecase.UseCase
}

func NewGroupCatalogHandler(useCase *scheduleusecase.UseCase) *GroupCatalogHandler {
	return &GroupCatalogHandler{useCase: useCase}
}

func (h *GroupCatalogHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	response, err := h.useCase.GetGroups(r.Context())
	if err != nil {
		httpjson.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to load groups")
		return
	}

	httpjson.WriteJSON(w, http.StatusOK, response)
}
