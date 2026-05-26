package memory

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
)

func TestCreateBookingRejectsOverlap(t *testing.T) {
	repository := New([]domain.Place{
		{
			ID:   "place_100",
			Name: "Аудитория 100",
			Type: "classroom",
			Coordinates: domain.Coordinates{
				Building: "1",
				Floor:    1,
			},
		},
	})

	startsAt := time.Date(2026, time.February, 9, 10, 0, 0, 0, time.UTC)
	request := domain.RoomBookingRequest{
		RoomID:        "place_100",
		StartsAt:      startsAt,
		EndsAt:        startsAt.Add(90 * time.Minute),
		BookerName:    "Student",
		BookerContact: "student@example.com",
	}

	if _, err := repository.CreateBooking(context.Background(), request); err != nil {
		t.Fatalf("expected first booking to succeed: %v", err)
	}
	if _, err := repository.CreateBooking(context.Background(), request); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}
}

func TestListAvailabilityExcludesBookedRoom(t *testing.T) {
	repository := New([]domain.Place{
		{
			ID:   "place_100",
			Name: "Аудитория 100",
			Type: "classroom",
			Coordinates: domain.Coordinates{
				Building: "1",
				Floor:    1,
			},
		},
	})

	startsAt := time.Date(2026, time.February, 9, 10, 0, 0, 0, time.UTC)
	if _, err := repository.CreateBooking(context.Background(), domain.RoomBookingRequest{
		RoomID:        "place_100",
		StartsAt:      startsAt,
		EndsAt:        startsAt.Add(90 * time.Minute),
		BookerName:    "Student",
		BookerContact: "student@example.com",
	}); err != nil {
		t.Fatalf("create booking: %v", err)
	}

	items, err := repository.ListAvailability(context.Background(), domain.RoomAvailabilityFilter{
		StartsAt: startsAt,
		EndsAt:   startsAt.Add(90 * time.Minute),
	})
	if err != nil {
		t.Fatalf("list availability: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected booked room to be excluded, got %d items", len(items))
	}
}

func TestCancelBookingReleasesRoom(t *testing.T) {
	repository := New([]domain.Place{
		{
			ID:   "place_100",
			Name: "Аудитория 100",
			Type: "classroom",
			Coordinates: domain.Coordinates{
				Building: "1",
				Floor:    1,
			},
		},
	})

	startsAt := time.Date(2026, time.February, 9, 10, 0, 0, 0, time.UTC)
	booking, err := repository.CreateBooking(context.Background(), domain.RoomBookingRequest{
		RoomID:        "place_100",
		StartsAt:      startsAt,
		EndsAt:        startsAt.Add(90 * time.Minute),
		BookerName:    "Student",
		BookerContact: "student@example.com",
	})
	if err != nil {
		t.Fatalf("create booking: %v", err)
	}

	if err := repository.CancelBooking(context.Background(), booking.ID); err != nil {
		t.Fatalf("cancel booking: %v", err)
	}

	items, err := repository.ListAvailability(context.Background(), domain.RoomAvailabilityFilter{
		StartsAt: startsAt,
		EndsAt:   startsAt.Add(90 * time.Minute),
	})
	if err != nil {
		t.Fatalf("list availability: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected room to become available, got %d items", len(items))
	}
}

func TestListBookingsReturnsBookingsForDate(t *testing.T) {
	repository := New([]domain.Place{
		{
			ID:   "place_100",
			Name: "Аудитория 100",
			Type: "classroom",
			Coordinates: domain.Coordinates{
				Building: "1",
				Floor:    1,
			},
		},
		{
			ID:   "place_200",
			Name: "Аудитория 200",
			Type: "classroom",
			Coordinates: domain.Coordinates{
				Building: "1",
				Floor:    2,
			},
		},
	})

	targetDate := time.Date(2026, time.May, 18, 0, 0, 0, 0, time.UTC)
	firstStart := targetDate.Add(10 * time.Hour)
	secondStart := targetDate.AddDate(0, 0, 1).Add(10 * time.Hour)

	if _, err := repository.CreateBooking(context.Background(), domain.RoomBookingRequest{
		RoomID:        "place_100",
		StartsAt:      firstStart,
		EndsAt:        firstStart.Add(90 * time.Minute),
		BookerName:    "Student",
		BookerContact: "student@example.com",
	}); err != nil {
		t.Fatalf("create target booking: %v", err)
	}
	if _, err := repository.CreateBooking(context.Background(), domain.RoomBookingRequest{
		RoomID:        "place_200",
		StartsAt:      secondStart,
		EndsAt:        secondStart.Add(90 * time.Minute),
		BookerName:    "Teacher",
		BookerContact: "teacher@example.com",
	}); err != nil {
		t.Fatalf("create other booking: %v", err)
	}

	items, err := repository.ListBookings(context.Background(), targetDate)
	if err != nil {
		t.Fatalf("list bookings: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one booking, got %d", len(items))
	}
	if items[0].RoomID != "place_100" {
		t.Fatalf("expected place_100 booking, got %s", items[0].RoomID)
	}
}
