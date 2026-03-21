package storage

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
)

type PlaceJSON struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Type         string          `json:"type"`
	Description  string          `json:"description"`
	Coordinates  CoordinatesJSON `json:"coordinates"`
	Tags         []string        `json:"tags"`
	IsAccessible bool            `json:"isAccessible"`
}

type CoordinatesJSON struct {
	Building string  `json:"building"`
	Floor    int     `json:"floor"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
}

type StubPlaceRepository struct {
	places []domain.Place
}

func NewStubPlaceRepository() *StubPlaceRepository {
	places, err := loadPlacesFromJSON("/GEO_Json/points.json", 5)
	if err != nil {
		// Fallback на дефубные данные при ошибке загрузки
		places = getFallbackPlaces()
	}

	return &StubPlaceRepository{
		places: places,
	}
}

func getFallbackPlaces() []domain.Place {
	return []domain.Place{
		{
			ID:          "wardrobe",
			Name:        "wardrobe",
			Type:        "classroom",
			Description: "",
			Coordinates: domain.Coordinates{
				Building: "УЛК",
				Floor:    1,
				X:        518,
				Y:        171,
			},
			Tags:         []string{},
			IsAccessible: true,
		},
		{
			ID:          "129y",
			Name:        "129y",
			Type:        "classroom",
			Description: "",
			Coordinates: domain.Coordinates{
				Building: "УЛК",
				Floor:    1,
				X:        460,
				Y:        218,
			},
			Tags:         []string{},
			IsAccessible: true,
		},
		{
			ID:          "127y",
			Name:        "127y",
			Type:        "classroom",
			Description: "",
			Coordinates: domain.Coordinates{
				Building: "УЛК",
				Floor:    1,
				X:        518,
				Y:        220,
			},
			Tags:         []string{},
			IsAccessible: true,
		},
		{
			ID:          "ladder",
			Name:        "ladder",
			Type:        "classroom",
			Description: "",
			Coordinates: domain.Coordinates{
				Building: "УЛК",
				Floor:    1,
				X:        428,
				Y:        218,
			},
			Tags:         []string{},
			IsAccessible: true,
		},
		{
			ID:          "toilet",
			Name:        "toilet",
			Type:        "classroom",
			Description: "",
			Coordinates: domain.Coordinates{
				Building: "УЛК",
				Floor:    1,
				X:        351,
				Y:        177,
			},
			Tags:         []string{},
			IsAccessible: true,
		},
	}
}

func loadPlacesFromJSON(filepath string, limit int) ([]domain.Place, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var placesJSON []PlaceJSON
	if err := json.Unmarshal(data, &placesJSON); err != nil {
		return nil, err
	}

	// Ограничиваем количество точек
	if len(placesJSON) > limit {
		placesJSON = placesJSON[:limit]
	}

	// Конвертируем в domain.Place
	places := make([]domain.Place, 0, len(placesJSON))
	for _, p := range placesJSON {
		places = append(places, domain.Place{
			ID:          p.ID,
			Name:        p.Name,
			Type:        p.Type,
			Description: p.Description,
			Coordinates: domain.Coordinates{
				Building: p.Coordinates.Building,
				Floor:    p.Coordinates.Floor,
				X:        p.Coordinates.X,
				Y:        p.Coordinates.Y,
			},
			Tags:         p.Tags,
			IsAccessible: p.IsAccessible,
		})
	}

	return places, nil
}

func (r *StubPlaceRepository) List(_ context.Context, filter domain.PlaceFilter) ([]domain.Place, error) {
	filtered := make([]domain.Place, 0, len(r.places))

	for _, place := range r.places {
		if filter.Type != "" && place.Type != filter.Type {
			continue
		}
		if filter.Building != "" && !strings.EqualFold(place.Coordinates.Building, filter.Building) {
			continue
		}
		if filter.Floor != nil && place.Coordinates.Floor != *filter.Floor {
			continue
		}
		if filter.Search != "" {
			query := strings.ToLower(filter.Search)
			if !strings.Contains(strings.ToLower(place.Name), query) &&
				!strings.Contains(strings.ToLower(place.Description), query) {
				continue
			}
		}

		filtered = append(filtered, place)
	}

	return filtered, nil
}

func (r *StubPlaceRepository) GetByID(_ context.Context, id string) (*domain.Place, error) {
	for _, place := range r.places {
		if place.ID == id {
			placeCopy := place
			return &placeCopy, nil
		}
	}

	return nil, domain.ErrNotFound
}
