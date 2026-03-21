package usecase

import (
	"context"
	"fmt"
	"math"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
)

type RouteUseCase struct {
	places domain.PlaceRepository
}

func NewRouteUseCase(places domain.PlaceRepository) *RouteUseCase {
	return &RouteUseCase{places: places}
}

func (u *RouteUseCase) Build(ctx context.Context, request domain.RouteRequest) (*domain.Route, error) {
	fromPlace, err := u.places.GetByID(ctx, request.FromPlaceID)
	if err != nil {
		return nil, err
	}

	toPlace, err := u.places.GetByID(ctx, request.ToPlaceID)
	if err != nil {
		return nil, err
	}

	distance := estimateDistance(fromPlace.Coordinates, toPlace.Coordinates)
	duration := distance / 45
	if duration == 0 {
		duration = 1
	}

	return &domain.Route{
		DistanceMeters:           distance,
		EstimatedDurationMinutes: duration,
		Steps: []domain.RouteStep{
			{
				Order:       1,
				Instruction: fmt.Sprintf("Начните маршрут от %s", fromPlace.Name),
				PlaceID:     fromPlace.ID,
				Coordinates: fromPlace.Coordinates,
			},
			{
				Order:       2,
				Instruction: fmt.Sprintf("Следуйте к точке %s", toPlace.Name),
				PlaceID:     toPlace.ID,
				Coordinates: toPlace.Coordinates,
			},
		},
	}, nil
}

func estimateDistance(from domain.Coordinates, to domain.Coordinates) int {
	if from.Building != to.Building {
		return 300
	}

	dx := from.X - to.X
	dy := from.Y - to.Y
	floorPenalty := math.Abs(float64(from.Floor-to.Floor)) * 25

	return int(math.Round(math.Sqrt(dx*dx+dy*dy) + floorPenalty))
}
