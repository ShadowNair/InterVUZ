package usecase

import (
	"context"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
)

type Repository interface {
	Get(ctx context.Context) (domain.NavigationGraph, error)
}

type UseCase struct {
	repository Repository
}

func New(repository Repository) *UseCase {
	return &UseCase{repository: repository}
}

func (uc *UseCase) Get(ctx context.Context) (domain.NavigationGraph, error) {
	return uc.repository.Get(ctx)
}
