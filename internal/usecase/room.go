package usecase

import (
	"context"
	"errors"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
)

type RoomUseCase struct {
	repo domain.RoomRepository
}

func NewRoomUseCase(repo domain.RoomRepository) *RoomUseCase {
	return &RoomUseCase{repo: repo}
}

func (u *RoomUseCase) ListAvailability(ctx context.Context, filter domain.RoomAvailabilityFilter) (*domain.RoomAvailabilityResponse, error) {
	if filter.StartsAt.IsZero() || filter.EndsAt.IsZero() {
		return nil, errors.New("startsAt and endsAt are required")
	}
	if !filter.StartsAt.Before(filter.EndsAt) {
		return nil, errors.New("startsAt must be before endsAt")
	}

	items, err := u.repo.ListAvailability(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &domain.RoomAvailabilityResponse{Items: items}, nil
}
