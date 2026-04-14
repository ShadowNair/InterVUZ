package schedulesync

import (
	"context"
	"fmt"
	"log"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
)

type GroupLister interface {
	ListGroupExternalUUIDs(ctx context.Context) ([]string, error)
}

type Source interface {
	FetchGroupSchedule(ctx context.Context, groupExternalUUID string) (*domain.GroupScheduleResponse, error)
}

type Repository interface {
	UpsertGroupSchedule(ctx context.Context, ownerGroupExternalUUID string, payload *domain.GroupScheduleResponse) (int, error)
}

type UseCase struct {
	groups GroupLister
	source Source
	repo   Repository
}

type SyncStats struct {
	GroupsTotal  int
	GroupsSynced int
	EventsSaved  int
}

func New(groups GroupLister, source Source, repo Repository) *UseCase {
	return &UseCase{groups: groups, source: source, repo: repo}
}

func (uc *UseCase) SyncAll(ctx context.Context) (*SyncStats, error) {
	groupUUIDs, err := uc.groups.ListGroupExternalUUIDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("list groups: %w", err)
	}

	stats := &SyncStats{GroupsTotal: len(groupUUIDs)}
	for _, groupUUID := range groupUUIDs {
		payload, err := uc.source.FetchGroupSchedule(ctx, groupUUID)
		if err != nil {
			log.Printf("schedule sync skipped for group %s: %v", groupUUID, err)
			continue
		}

		saved, err := uc.repo.UpsertGroupSchedule(ctx, groupUUID, payload)
		if err != nil {
			log.Printf("schedule save skipped for group %s: %v", groupUUID, err)
			continue
		}

		stats.GroupsSynced++
		stats.EventsSaved += saved
	}

	return stats, nil
}
