package usecase

import (
	"context"
	"errors"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
)

var ErrNotFound = errors.New("place not found")

type Repository interface {
	List(ctx context.Context, filter domain.PlaceFilter) ([]domain.Place, error)
	GetByID(ctx context.Context, placeID string) (domain.Place, error)
}

type UseCase struct {
	repository Repository
}

func New(repository Repository) *UseCase {
	return &UseCase{repository: repository}
}

func (uc *UseCase) List(ctx context.Context, filter domain.PlaceFilter) (*domain.PlacesResponse, error) {
	items, err := uc.repository.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &domain.PlacesResponse{
		Items: items,
		Total: len(items),
	}, nil
}

func (uc *UseCase) GetByID(ctx context.Context, placeID string) (domain.Place, error) {
	place, err := uc.repository.GetByID(ctx, placeID)
	if err == nil {
		return place, nil
	}

	if errors.Is(err, domain.ErrNotFound) {
		return domain.Place{}, ErrNotFound
	}

	return domain.Place{}, err
}
