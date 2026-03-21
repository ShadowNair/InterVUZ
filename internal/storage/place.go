package storage

import (
	"context"
	"strings"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
)

type StubPlaceRepository struct {
	places []domain.Place
}

func NewStubPlaceRepository() *StubPlaceRepository {
	return &StubPlaceRepository{
		places: []domain.Place{
			{
				ID:          "place_101",
				Name:        "Аудитория 214",
				Type:        "classroom",
				Description: "Компьютерный класс кафедры информатики",
				Coordinates: domain.Coordinates{
					Building: "B1",
					Floor:    2,
					X:        18.5,
					Y:        42.0,
				},
				Tags:         []string{"computer", "projector"},
				IsAccessible: true,
			},
			{
				ID:          "place_102",
				Name:        "Столовая главного корпуса",
				Type:        "cafeteria",
				Description: "Основная столовая для студентов и сотрудников",
				Coordinates: domain.Coordinates{
					Building: "B1",
					Floor:    1,
					X:        7.0,
					Y:        15.0,
				},
				Tags:         []string{"food", "coffee"},
				IsAccessible: true,
			},
			{
				ID:          "place_103",
				Name:        "Принтер 2 этаж",
				Type:        "printer",
				Description: "Точка печати рядом с деканатом",
				Coordinates: domain.Coordinates{
					Building: "B1",
					Floor:    2,
					X:        20.0,
					Y:        30.0,
				},
				Tags:         []string{"print", "documents"},
				IsAccessible: true,
			},
			{
				ID:          "place_104",
				Name:        "Деканат ИУ",
				Type:        "dean_office",
				Description: "Деканат факультета информатики и управления",
				Coordinates: domain.Coordinates{
					Building: "B2",
					Floor:    3,
					X:        12.0,
					Y:        11.0,
				},
				Tags:         []string{"office"},
				IsAccessible: false,
			},
		},
	}
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
