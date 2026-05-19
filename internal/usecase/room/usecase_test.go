package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
)

type repositoryStub struct {
	availabilityFilter domain.RoomAvailabilityFilter
}

func (stub *repositoryStub) ListAvailability(_ context.Context, filter domain.RoomAvailabilityFilter) ([]domain.RoomAvailability, error) {
	stub.availabilityFilter = filter
	return nil, nil
}

func (stub *repositoryStub) GetSchedule(context.Context, string, time.Time) (*domain.RoomScheduleResponse, error) {
	return nil, nil
}

func (stub *repositoryStub) ListBookings(context.Context, time.Time) ([]domain.RoomBooking, error) {
	return nil, nil
}

func (stub *repositoryStub) CreateBooking(context.Context, domain.RoomBookingRequest) (*domain.RoomBooking, error) {
	return nil, nil
}

func (stub *repositoryStub) CancelBooking(context.Context, string) error {
	return nil
}

func TestListAvailabilityDefaultsToBookingDuration(t *testing.T) {
	repository := &repositoryStub{}
	useCase := New(repository)

	startsAt := time.Date(2026, time.February, 9, 10, 0, 0, 0, time.UTC)
	if _, err := useCase.ListAvailability(context.Background(), domain.RoomAvailabilityFilter{
		StartsAt: startsAt,
	}); err != nil {
		t.Fatalf("expected availability response, got error: %v", err)
	}

	if got, want := repository.availabilityFilter.EndsAt, startsAt.Add(BookingDuration); !got.Equal(want) {
		t.Fatalf("expected endsAt %s, got %s", want, got)
	}
}

func TestCreateBookingValidation(t *testing.T) {
	useCase := New(&repositoryStub{})

	_, err := useCase.CreateBooking(context.Background(), domain.RoomBookingRequest{
		RoomID:        "place_100",
		StartsAt:      time.Date(2026, time.February, 9, 10, 0, 0, 0, time.UTC),
		BookerName:    " ",
		BookerContact: "student@example.com",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}
