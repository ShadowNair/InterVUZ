package memory

import (
	"context"
	"strings"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
)

type Repository struct {
	places     []domain.Place
	placesByID map[string]domain.Place
}

func New(places []domain.Place) *Repository {
	items := make([]domain.Place, 0, len(places))
	index := make(map[string]domain.Place, len(places))

	for _, place := range places {
		placeCopy := place
		placeCopy.Tags = append([]string(nil), place.Tags...)
		items = append(items, placeCopy)
		index[placeCopy.ID] = placeCopy
	}

	return &Repository{
		places:     items,
		placesByID: index,
	}
}

func (r *Repository) List(_ context.Context, filter domain.PlaceFilter) ([]domain.Place, error) {
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

		placeCopy := place
		placeCopy.Tags = append([]string(nil), place.Tags...)
		filtered = append(filtered, placeCopy)
	}

	return filtered, nil
}

func (r *Repository) GetByID(_ context.Context, placeID string) (domain.Place, error) {
	place, ok := r.placesByID[placeID]
	if !ok {
		return domain.Place{}, domain.ErrNotFound
	}

	place.Tags = append([]string(nil), place.Tags...)
	return place, nil
}
