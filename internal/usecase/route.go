package usecase

import (
	"context"
	"fmt"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
)

type RouteUseCase struct {
	places domain.PlaceRepository
}

func NewRouteUseCase(places domain.PlaceRepository) *RouteUseCase {
	return &RouteUseCase{places: places}
}

func (u *RouteUseCase) Build(ctx context.Context, request domain.RouteRequest) (*domain.Route, error) {
	places, err := u.places.List(ctx, domain.PlaceFilter{})
	if err != nil {
		return nil, err
	}

	//if request.AccessibleOnly {
	//	places = filterAccessiblePlaces(places)
	//}

	graph := domain.NewPlaceGraph(places)
	if graph == nil {
		return nil, domain.ErrNotFound
	}

	if err := graph.ConnectAllNodes(); err != nil {
		return nil, err
	}

	path, err := graph.FindShortestWay(request.FromPlaceID, request.ToPlaceID)
	if err != nil {
		return nil, err
	}

	distance := calculatePathDistance(path)
	duration := distance / 45
	if duration == 0 {
		duration = 1
	}

	steps := buildRouteSteps(path)

	return &domain.Route{
		DistanceMeters:           distance,
		EstimatedDurationMinutes: duration,
		Steps:                    steps,
	}, nil
}

func filterAccessiblePlaces(places []domain.Place) []domain.Place {
	filtered := make([]domain.Place, 0, len(places))

	for _, place := range places {
		if place.IsAccessible {
			filtered = append(filtered, place)
		}
	}

	return filtered
}

func calculatePathDistance(path []domain.Place) int {
	if len(path) < 2 {
		return 0
	}

	total := 0
	for i := 1; i < len(path); i++ {
		total += domain.EstimateDistance(path[i-1].Coordinates, path[i].Coordinates)
	}

	return total
}

func buildRouteSteps(path []domain.Place) []domain.RouteStep {
	steps := make([]domain.RouteStep, 0, len(path))
	for i, place := range path {
		steps = append(steps, domain.RouteStep{
			Order:       i + 1,
			Instruction: buildStepInstruction(path, i),
			PlaceID:     place.ID,
			Coordinates: place.Coordinates,
		})
	}

	return steps
}

func buildStepInstruction(path []domain.Place, index int) string {
	place := path[index]

	if index == 0 {
		return fmt.Sprintf("Начните маршрут от %s", place.Name)
	}

	if index == len(path)-1 {
		return fmt.Sprintf("Следуйте к точке %s", place.Name)
	}

	return fmt.Sprintf("Продолжайте маршрут через %s", place.Name)
}
