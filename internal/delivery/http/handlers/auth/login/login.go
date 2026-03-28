package login

import (
	"encoding/json"
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
	if r.URL.Path != "/auth/login" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		httputil.WriteMethodNotAllowed(w)
		return
	}

	var request domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "bad_request", "invalid json body")
		return
	}

	response, err := h.useCase.Login(r.Context(), request)
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
