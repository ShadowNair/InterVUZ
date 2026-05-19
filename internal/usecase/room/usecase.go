package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
)

const BookingDuration = 90 * time.Minute

type Repository interface {
	ListAvailability(ctx context.Context, filter domain.RoomAvailabilityFilter) ([]domain.RoomAvailability, error)
	GetSchedule(ctx context.Context, roomID string, date time.Time) (*domain.RoomScheduleResponse, error)
	ListBookings(ctx context.Context, date time.Time) ([]domain.RoomBooking, error)
	CreateBooking(ctx context.Context, request domain.RoomBookingRequest) (*domain.RoomBooking, error)
	CancelBooking(ctx context.Context, bookingID string) error
}

type UseCase struct {
	repository Repository
}

func New(repository Repository) *UseCase {
	return &UseCase{repository: repository}
}

func (uc *UseCase) ListAvailability(ctx context.Context, filter domain.RoomAvailabilityFilter) (*domain.RoomAvailabilityResponse, error) {
	if filter.StartsAt.IsZero() {
		return nil, errors.New("startsAt is required")
	}
	if filter.EndsAt.IsZero() {
		filter.EndsAt = filter.StartsAt.Add(BookingDuration)
	}
	if !filter.StartsAt.Before(filter.EndsAt) {
		return nil, errors.New("startsAt must be before endsAt")
	}

	items, err := uc.repository.ListAvailability(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &domain.RoomAvailabilityResponse{Items: items}, nil
}

func (uc *UseCase) GetSchedule(ctx context.Context, roomID string, date time.Time) (*domain.RoomScheduleResponse, error) {
	if strings.TrimSpace(roomID) == "" {
		return nil, errors.New("room id is required")
	}
	if date.IsZero() {
		return nil, errors.New("date is required")
	}

	return uc.repository.GetSchedule(ctx, strings.TrimSpace(roomID), date)
}

func (uc *UseCase) ListBookings(ctx context.Context, date time.Time) (*domain.RoomBookingsResponse, error) {
	if date.IsZero() {
		return nil, errors.New("date is required")
	}

	items, err := uc.repository.ListBookings(ctx, date)
	if err != nil {
		return nil, err
	}

	return &domain.RoomBookingsResponse{
		Date:  date.Format("2006-01-02"),
		Items: items,
	}, nil
}

func (uc *UseCase) CreateBooking(ctx context.Context, request domain.RoomBookingRequest) (*domain.RoomBooking, error) {
	request.RoomID = strings.TrimSpace(request.RoomID)
	request.BookerName = strings.TrimSpace(request.BookerName)
	request.BookerContact = strings.TrimSpace(request.BookerContact)

	if request.RoomID == "" {
		return nil, errors.New("room id is required")
	}
	if request.StartsAt.IsZero() {
		return nil, errors.New("startsAt is required")
	}
	if request.BookerName == "" {
		return nil, errors.New("bookerName is required")
	}
	if request.BookerContact == "" {
		return nil, errors.New("bookerContact is required")
	}

	request.EndsAt = request.StartsAt.Add(BookingDuration)

	return uc.repository.CreateBooking(ctx, request)
}

func (uc *UseCase) CancelBooking(ctx context.Context, bookingID string) error {
	bookingID = strings.TrimSpace(bookingID)
	if bookingID == "" {
		return errors.New("booking id is required")
	}

	return uc.repository.CancelBooking(ctx, bookingID)
}
