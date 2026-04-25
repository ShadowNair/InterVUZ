package schedulesync

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
)

type GroupLister interface {
	ListGroupExternalUUIDs(ctx context.Context) ([]string, error)
}

type Source interface {
	FetchGroupSchedule(ctx context.Context, groupExternalUUID string) (*domain.GroupScheduleResponse, error)
}

type Repository interface {
	ReplaceScheduledGroups(ctx context.Context, checkedGroupExternalUUIDs []string, scheduledGroupExternalUUIDs []string) error
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
	checkedGroupUUIDs := make([]string, 0, len(groupUUIDs))
	scheduledPayloads := make([]groupSchedulePayload, 0, len(groupUUIDs))

	for _, groupUUID := range groupUUIDs {
		payload, err := uc.source.FetchGroupSchedule(ctx, groupUUID)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				checkedGroupUUIDs = append(checkedGroupUUIDs, groupUUID)
				log.Printf("schedule sync skipped for group %s: schedule not found", groupUUID)
				continue
			}

			log.Printf("schedule sync skipped for group %s: %v", groupUUID, err)
			continue
		}

		checkedGroupUUIDs = append(checkedGroupUUIDs, groupUUID)
		if !hasSchedule(payload) {
			log.Printf("schedule sync skipped for group %s: empty schedule", groupUUID)
			continue
		}

		scheduledPayloads = append(scheduledPayloads, groupSchedulePayload{
			groupExternalUUID: groupUUID,
			payload:           payload,
		})
	}

	scheduledGroupUUIDs := make([]string, 0, len(scheduledPayloads))
	for _, item := range scheduledPayloads {
		scheduledGroupUUIDs = append(scheduledGroupUUIDs, item.groupExternalUUID)
	}

	if err := uc.repo.ReplaceScheduledGroups(ctx, checkedGroupUUIDs, scheduledGroupUUIDs); err != nil {
		return nil, fmt.Errorf("replace scheduled groups: %w", err)
	}

	for _, item := range scheduledPayloads {
		saved, err := uc.repo.UpsertGroupSchedule(ctx, item.groupExternalUUID, item.payload)
		if err != nil {
			log.Printf("schedule save skipped for group %s: %v", item.groupExternalUUID, err)
			continue
		}

		stats.GroupsSynced++
		stats.EventsSaved += saved
	}

	return stats, nil
}

type groupSchedulePayload struct {
	groupExternalUUID string
	payload           *domain.GroupScheduleResponse
}

func hasSchedule(payload *domain.GroupScheduleResponse) bool {
	if payload == nil {
		return false
	}

	if len(payload.Data.Schedule) == 0 {
		return false
	}

	for _, event := range payload.Data.Schedule {
		if event.Day > 0 || event.Time > 0 || strings.TrimSpace(event.StartTime) != "" || strings.TrimSpace(event.EndTime) != "" {
			return true
		}
	}

	return false
}
