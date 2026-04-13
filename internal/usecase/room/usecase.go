package usecase

import (
	"context"
	"errors"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
)

type Repository interface {
	ListAvailability(ctx context.Context, filter domain.RoomAvailabilityFilter) ([]domain.RoomAvailability, error)
}

type UseCase struct {
	repository Repository
}

func New(repository Repository) *UseCase {
	return &UseCase{repository: repository}
}

func (uc *UseCase) ListAvailability(ctx context.Context, filter domain.RoomAvailabilityFilter) (*domain.RoomAvailabilityResponse, error) {
	if filter.StartsAt.IsZero() || filter.EndsAt.IsZero() {
		return nil, errors.New("startsAt and endsAt are required")
	}
	if !filter.StartsAt.Before(filter.EndsAt) {
		return nil, errors.New("startsAt must be before endsAt")
	}

	items, err := uc.repository.ListAvailability(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &domain.RoomAvailabilityResponse{Items: items}, nil
}
