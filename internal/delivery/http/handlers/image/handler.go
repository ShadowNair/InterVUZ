package image

import (
	"errors"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"os"
	"strconv"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/handlers/common"
)

type Handler struct {
	filePath string
}

func NewHandler(filePath string) *Handler {
	return &Handler{filePath: filePath}
}

func (h *Handler) Serve(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/image" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		common.WriteMethodNotAllowed(w)
		return
	}

	if _, err := os.Stat(h.filePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			common.WriteError(w, http.StatusNotFound, "not_found", "image file not found")
			return
		}

		common.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to read image file")
		return
	}

	width, height, err := readImageSize(h.filePath)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to decode image size")
		return
	}

	w.Header().Set("X-Image-Width", strconv.Itoa(width))
	w.Header().Set("X-Image-Height", strconv.Itoa(height))

	http.ServeFile(w, r, h.filePath)
}

func readImageSize(path string) (int, int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	config, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, err
	}

	return config.Width, config.Height, nil
}
