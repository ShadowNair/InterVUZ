package me

import (
	"net/http"

	authcommon "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/handlers/auth/common"
	httputil "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/handlers/common"
	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/usecase/profile"
)

type Handler struct {
	useCase *profile.ProfileUCase
}

func NewHandler(useCase *profile.ProfileUCase) *Handler {
	return &Handler{useCase: useCase}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/auth/me" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		httputil.WriteMethodNotAllowed(w)
		return
	}

	accessToken, err := authcommon.ExtractAccessToken(r)
	if err != nil {
		authcommon.WriteUseCaseError(w, err)
		return
	}

	response, err := h.useCase.Me(r.Context(), accessToken)
	if err != nil {
		authcommon.WriteUseCaseError(w, err)
		return
	}

	response.Code = http.StatusOK
	httputil.WriteJSON(w, http.StatusOK, response)
}
