package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/platform/httpjson"
	assistantusecase "github.com/GIT_USER_ID/GIT_REPO_ID/internal/usecase/assistant"
)

type ChatHandler struct {
	useCase *assistantusecase.UseCase
}

func NewChatHandler(useCase *assistantusecase.UseCase) *ChatHandler {
	return &ChatHandler{useCase: useCase}
}

func (h *ChatHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var request domain.AssistantChatRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httpjson.WriteError(w, http.StatusBadRequest, "bad_request", "invalid json body")
		return
	}

	response, err := h.useCase.Chat(r.Context(), request)
	if err != nil {
		switch {
		case errors.Is(err, assistantusecase.ErrInvalidRequest):
			httpjson.WriteError(w, http.StatusBadRequest, "bad_request", err.Error())
		case errors.Is(err, assistantusecase.ErrUnavailable):
			httpjson.WriteError(w, http.StatusServiceUnavailable, "assistant_unavailable", "assistant service is unavailable")
		default:
			httpjson.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to process assistant request")
		}
		return
	}

	httpjson.WriteJSON(w, http.StatusOK, response)
}
