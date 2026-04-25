package news

import (
	"context"
	"fmt"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
)

type Source interface {
	LoadNews(ctx context.Context, limit int) ([]domain.NewsItem, error)
}

type Repository interface {
	SaveMany(ctx context.Context, items []domain.NewsRecord) error
	List(ctx context.Context, limit int) ([]domain.NewsRecord, error)
}

type UseCase struct {
	source Source
	repo   Repository
}

func New(source Source, repo Repository) *UseCase {
	return &UseCase{
		source: source,
		repo:   repo,
	}
}

func (u *UseCase) Sync(ctx context.Context, limit int) error {
	items, err := u.source.LoadNews(ctx, limit)
	if err != nil {
		return fmt.Errorf("load news: %w", err)
	}

	records := make([]domain.NewsRecord, 0, len(items))
	for _, item := range items {
		record, err := ToNewsRecord(item)
		if err != nil {
			return fmt.Errorf("map news item %s: %w", item.Slug, err)
		}
		records = append(records, record)
	}

	if err := u.repo.SaveMany(ctx, records); err != nil {
		return fmt.Errorf("save news: %w", err)
	}

	return nil
}

func (u *UseCase) List(ctx context.Context, limit int) (*domain.NewsResponse, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	records, err := u.repo.List(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("list news: %w", err)
	}

	items := make([]domain.NewsItem, 0, len(records))
	for _, record := range records {
		items = append(items, ToNewsItem(record))
	}

	return &domain.NewsResponse{Items: items}, nil
}
