package logout

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
	if r.URL.Path != "/auth/logout" {
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

	if err := h.useCase.Logout(r.Context(), domain.LogoutRequest{RefreshToken: refreshToken}); err != nil {
		authcommon.WriteUseCaseError(w, err)
		return
	}

	authcommon.ClearAccessTokenCookie(w)
	authcommon.ClearRefreshTokenCookie(w)
	httputil.WriteJSON(w, http.StatusOK, domain.CodeResponse{Code: http.StatusOK})
}
