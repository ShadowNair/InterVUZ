package schedulesync

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
)

func TestSyncAllSavesOnlyGroupsWithSchedules(t *testing.T) {
	ctx := context.Background()
	sourceErr := errors.New("temporary source error")

	uc := New(
		fakeGroupLister{groups: []string{"scheduled", "empty", "missing", "flaky"}},
		fakeScheduleSource{
			payloads: map[string]*domain.GroupScheduleResponse{
				"scheduled": {
					Data: domain.GroupSchedule{
						UUID: "scheduled",
						Schedule: []domain.ScheduleEventItem{
							{Day: 1, Time: 1, StartTime: "08:30", EndTime: "10:00"},
							{Day: 2, Time: 2, StartTime: "10:10", EndTime: "11:40"},
						},
					},
				},
				"empty": {
					Data: domain.GroupSchedule{
						UUID:     "empty",
						Schedule: []domain.ScheduleEventItem{},
					},
				},
			},
			errors: map[string]error{
				"missing": domain.ErrNotFound,
				"flaky":   sourceErr,
			},
		},
		&fakeScheduleRepo{},
	)

	repo := uc.repo.(*fakeScheduleRepo)
	stats, err := uc.SyncAll(ctx)
	if err != nil {
		t.Fatalf("SyncAll returned error: %v", err)
	}

	if stats.GroupsTotal != 4 {
		t.Fatalf("GroupsTotal = %d, want 4", stats.GroupsTotal)
	}
	if stats.GroupsSynced != 1 {
		t.Fatalf("GroupsSynced = %d, want 1", stats.GroupsSynced)
	}
	if stats.EventsSaved != 2 {
		t.Fatalf("EventsSaved = %d, want 2", stats.EventsSaved)
	}

	if want := []string{"scheduled", "empty", "missing"}; !reflect.DeepEqual(repo.checkedGroups, want) {
		t.Fatalf("checked groups = %#v, want %#v", repo.checkedGroups, want)
	}
	if want := []string{"scheduled"}; !reflect.DeepEqual(repo.scheduledGroups, want) {
		t.Fatalf("scheduled groups = %#v, want %#v", repo.scheduledGroups, want)
	}
	if want := []string{"scheduled"}; !reflect.DeepEqual(repo.upsertedGroups, want) {
		t.Fatalf("upserted groups = %#v, want %#v", repo.upsertedGroups, want)
	}
}

type fakeGroupLister struct {
	groups []string
	err    error
}

func (f fakeGroupLister) ListGroupExternalUUIDs(context.Context) ([]string, error) {
	return f.groups, f.err
}

type fakeScheduleSource struct {
	payloads map[string]*domain.GroupScheduleResponse
	errors   map[string]error
}

func (f fakeScheduleSource) FetchGroupSchedule(_ context.Context, groupExternalUUID string) (*domain.GroupScheduleResponse, error) {
	if err := f.errors[groupExternalUUID]; err != nil {
		return nil, err
	}
	return f.payloads[groupExternalUUID], nil
}

type fakeScheduleRepo struct {
	checkedGroups   []string
	scheduledGroups []string
	upsertedGroups  []string
}

func (f *fakeScheduleRepo) ReplaceScheduledGroups(_ context.Context, checkedGroupExternalUUIDs []string, scheduledGroupExternalUUIDs []string) error {
	f.checkedGroups = append([]string(nil), checkedGroupExternalUUIDs...)
	f.scheduledGroups = append([]string(nil), scheduledGroupExternalUUIDs...)
	return nil
}

func (f *fakeScheduleRepo) UpsertGroupSchedule(_ context.Context, ownerGroupExternalUUID string, payload *domain.GroupScheduleResponse) (int, error) {
	f.upsertedGroups = append(f.upsertedGroups, ownerGroupExternalUUID)
	return len(payload.Data.Schedule), nil
}
