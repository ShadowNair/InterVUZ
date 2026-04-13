package structure

import (
	"context"
	"fmt"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
)

type Source interface {
	FetchUnits(ctx context.Context) ([]domain.StructureUnit, error)
}

type Repository interface {
	UpsertMany(ctx context.Context, items []domain.StructureUnit) error
}

type UseCase struct {
	source Source
	repo   Repository
}

func New(source Source, repo Repository) *UseCase {
	return &UseCase{source: source, repo: repo}
}

func (uc *UseCase) Sync(ctx context.Context) (int, error) {
	items, err := uc.source.FetchUnits(ctx)
	if err != nil {
		return 0, fmt.Errorf("fetch structure units: %w", err)
	}

	if len(items) == 0 {
		return 0, nil
	}

	if err := uc.repo.UpsertMany(ctx, items); err != nil {
		return 0, fmt.Errorf("save structure units: %w", err)
	}

	return len(items), nil
}
