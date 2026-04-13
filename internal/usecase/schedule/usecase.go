package usecase

import (
	"context"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
)

type Repository interface {
	Import(ctx context.Context, request domain.ScheduleImportRequest) (*domain.ScheduleImportResult, error)
	GetGroupCatalog(ctx context.Context) (*domain.GroupCatalogResponse, error)
	GetGroupSchedule(ctx context.Context, groupID string) (*domain.GroupScheduleResponse, error)
	GetEvent(ctx context.Context, eventID string) (*domain.ScheduleEventResponse, error)
}

type UseCase struct {
	repository Repository
}

func New(repository Repository) *UseCase {
	return &UseCase{repository: repository}
}

func (uc *UseCase) Import(ctx context.Context, request domain.ScheduleImportRequest) (*domain.ScheduleImportResult, error) {
	return uc.repository.Import(ctx, request)
}

func (uc *UseCase) GetGroups(ctx context.Context) (*domain.GroupCatalogResponse, error) {
	return uc.repository.GetGroupCatalog(ctx)
}

func (uc *UseCase) GetGroupSchedule(ctx context.Context, groupID string) (*domain.GroupScheduleResponse, error) {
	return uc.repository.GetGroupSchedule(ctx, groupID)
}

func (uc *UseCase) GetEvent(ctx context.Context, eventID string) (*domain.ScheduleEventResponse, error) {
	return uc.repository.GetEvent(ctx, eventID)
}
