package usecase

import (
	"context"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
)

type ScheduleUseCase struct {
	repo domain.ScheduleRepository
}

func NewScheduleUseCase(repo domain.ScheduleRepository) *ScheduleUseCase {
	return &ScheduleUseCase{repo: repo}
}

func (u *ScheduleUseCase) Import(ctx context.Context, request domain.ScheduleImportRequest) (*domain.ScheduleImportResult, error) {
	return u.repo.Import(ctx, request)
}

func (u *ScheduleUseCase) GetGroups(ctx context.Context) (*domain.GroupCatalogResponse, error) {
	return u.repo.GetGroupCatalog(ctx)
}

func (u *ScheduleUseCase) GetGroupSchedule(ctx context.Context, groupID string) (*domain.GroupScheduleResponse, error) {
	return u.repo.GetGroupSchedule(ctx, groupID)
}

func (u *ScheduleUseCase) GetEvent(ctx context.Context, eventID string) (*domain.ScheduleEventResponse, error) {
	return u.repo.GetEvent(ctx, eventID)
}
