package places

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/handlers/common"
	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/usecase"
)

type Handler struct {
	useCase *usecase.PlaceUseCase
}

func NewHandler(useCase *usecase.PlaceUseCase) *Handler {
	return &Handler{useCase: useCase}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/places" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		common.WriteMethodNotAllowed(w)
		return
	}

	filter := domain.PlaceFilter{
		Type:     r.URL.Query().Get("type"),
		Building: r.URL.Query().Get("building"),
		Search:   r.URL.Query().Get("search"),
	}

	if floorRaw := r.URL.Query().Get("floor"); floorRaw != "" {
		floor, err := strconv.Atoi(floorRaw)
		if err != nil {
			common.WriteError(w, http.StatusBadRequest, "bad_request", "floor must be an integer")
			return
		}
		filter.Floor = &floor
	}

	response, err := h.useCase.List(r.Context(), filter)
	if err != nil {
		common.WriteDomainError(w, err)
		return
	}

	common.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		common.WriteMethodNotAllowed(w)
		return
	}

	placeID := strings.TrimPrefix(r.URL.Path, "/places/")
	if placeID == "" || strings.Contains(placeID, "/") {
		http.NotFound(w, r)
		return
	}

	response, err := h.useCase.GetByID(r.Context(), placeID)
	if err != nil {
		common.WriteDomainError(w, err)
		return
	}

	common.WriteJSON(w, http.StatusOK, response)
}
