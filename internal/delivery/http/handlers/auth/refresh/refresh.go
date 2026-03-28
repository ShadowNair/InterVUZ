package refresh

import (
	"net/http"

	authcommon "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/handlers/auth/common"
	httputil "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/handlers/common"
	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/usecase/profile"
)

type Handler struct {
	useCase *profile.ProfileUCase
}

func NewHandler(useCase *profile.ProfileUCase) *Handler {
	return &Handler{useCase: useCase}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/auth/refresh" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		httputil.WriteMethodNotAllowed(w)
		return
	}

	refreshToken, err := authcommon.ExtractRefreshTokenCookie(r)
	if err != nil {
		authcommon.WriteUseCaseError(w, err)
		return
	}

	response, err := h.useCase.Refresh(r.Context(), domain.RefreshTokenRequest{RefreshToken: refreshToken})
	if err != nil {
		authcommon.WriteUseCaseError(w, err)
		return
	}

	authcommon.SetAccessTokenCookie(w, response)
	authcommon.SetRefreshTokenCookie(w, response)
	payload := h.useCase.ToAccessTokenResponse(response)
	payload.Code = http.StatusOK
	httputil.WriteJSON(w, http.StatusOK, payload)
}
