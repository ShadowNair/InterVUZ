package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
)

type StubScheduleRepository struct {
	groupCatalog domain.GroupCatalogResponse
	groupData    map[string]domain.GroupScheduleResponse
	eventData    map[string]domain.ScheduleEventResponse
}

func NewStubScheduleRepository(dataDir string) (*StubScheduleRepository, error) {
	groupCatalogPath := filepath.Join(dataDir, "schedule_ID.json")
	groupSchedulePath := filepath.Join(dataDir, "schedule_group.json")

	groupCatalogBytes, err := os.ReadFile(groupCatalogPath)
	if err != nil {
		return nil, fmt.Errorf("read group catalog stub: %w", err)
	}

	groupScheduleBytes, err := os.ReadFile(groupSchedulePath)
	if err != nil {
		return nil, fmt.Errorf("read group schedule stub: %w", err)
	}

	var catalog domain.GroupCatalogResponse
	if err := json.Unmarshal(groupCatalogBytes, &catalog); err != nil {
		return nil, fmt.Errorf("decode group catalog stub: %w", err)
	}

	var groupSchedule domain.GroupScheduleResponse
	if err := json.Unmarshal(groupScheduleBytes, &groupSchedule); err != nil {
		return nil, fmt.Errorf("decode group schedule stub: %w", err)
	}

	repo := &StubScheduleRepository{
		groupCatalog: catalog,
		groupData: map[string]domain.GroupScheduleResponse{
			groupSchedule.Data.UUID: groupSchedule,
		},
		eventData: make(map[string]domain.ScheduleEventResponse),
	}

	for index, event := range groupSchedule.Data.Schedule {
		eventID := fmt.Sprintf("event-%d", index+1)
		repo.eventData[eventID] = domain.ScheduleEventResponse{Data: event}
	}

	return repo, nil
}

func (r *StubScheduleRepository) Import(_ context.Context, request domain.ScheduleImportRequest) (*domain.ScheduleImportResult, error) {
	return &domain.ScheduleImportResult{
		ImportID:       "import-stub-001",
		Status:         "queued",
		ImportedEvents: len(request.Events),
	}, nil
}

func (r *StubScheduleRepository) GetGroupCatalog(_ context.Context) (*domain.GroupCatalogResponse, error) {
	response := r.groupCatalog
	return &response, nil
}

func (r *StubScheduleRepository) GetGroupSchedule(_ context.Context, groupID string) (*domain.GroupScheduleResponse, error) {
	schedule, ok := r.groupData[groupID]
	if !ok {
		return nil, domain.ErrNotFound
	}

	response := schedule
	return &response, nil
}

func (r *StubScheduleRepository) GetEvent(_ context.Context, eventID string) (*domain.ScheduleEventResponse, error) {
	event, ok := r.eventData[eventID]
	if !ok {
		return nil, domain.ErrNotFound
	}

	response := event
	return &response, nil
}
