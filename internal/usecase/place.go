package usecase

import (
	"context"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
)

type PlaceUseCase struct {
	repo domain.PlaceRepository
}

func NewPlaceUseCase(repo domain.PlaceRepository) *PlaceUseCase {
	return &PlaceUseCase{repo: repo}
}

func (u *PlaceUseCase) List(ctx context.Context, filter domain.PlaceFilter) (*domain.PlacesResponse, error) {
	items, err := u.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &domain.PlacesResponse{
		Items: items,
		Total: len(items),
	}, nil
}

func (u *PlaceUseCase) GetByID(ctx context.Context, id string) (*domain.Place, error) {
	return u.repo.GetByID(ctx, id)
}
