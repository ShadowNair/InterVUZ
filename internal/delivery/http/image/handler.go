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

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/platform/httpjson"
)

type Handler struct {
	filePath string
}

func NewHandler(filePath string) *Handler {
	return &Handler{filePath: filePath}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if _, err := os.Stat(h.filePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			httpjson.WriteError(w, http.StatusNotFound, "not_found", "image file not found")
			return
		}

		httpjson.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to read image file")
		return
	}

	width, height, err := readImageSize(h.filePath)
	if err != nil {
		httpjson.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to decode image size")
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
