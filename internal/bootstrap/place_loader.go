package bootstrap

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
)

type placeSeedRecord struct {
	ID           string              `json:"id"`
	Name         string              `json:"name"`
	Type         string              `json:"type"`
	Description  string              `json:"description"`
	Coordinates  placeSeedCoordinate `json:"coordinates"`
	Tags         []string            `json:"tags"`
	IsAccessible bool                `json:"isAccessible"`
}

type placeSeedCoordinate struct {
	Building string  `json:"building"`
	Floor    int     `json:"floor"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
}

func LoadPlacesSeed(path string) ([]domain.Place, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read places seed %s: %w", filepath.Clean(path), err)
	}

	var records []placeSeedRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("decode places seed %s: %w", filepath.Clean(path), err)
	}

	places := make([]domain.Place, 0, len(records))
	for _, record := range records {
		places = append(places, domain.Place{
			ID:          strings.TrimSpace(record.ID),
			Name:        strings.TrimSpace(record.Name),
			Type:        strings.TrimSpace(record.Type),
			Description: strings.TrimSpace(record.Description),
			Coordinates: domain.Coordinates{
				Building: strings.TrimSpace(record.Coordinates.Building),
				Floor:    record.Coordinates.Floor,
				X:        record.Coordinates.X,
				Y:        record.Coordinates.Y,
			},
			Tags:         append([]string(nil), record.Tags...),
			IsAccessible: record.IsAccessible,
		})
	}

	return places, nil
}
