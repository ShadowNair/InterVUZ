package memory

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
)

type Repository struct {
	mu         sync.Mutex
	rooms      []domain.Place
	roomsByID  map[string]domain.Place
	bookings   []storedBooking
	nextNumber int
}

type storedBooking struct {
	id      string
	request domain.RoomBookingRequest
}

func New(places []domain.Place) *Repository {
	rooms := make([]domain.Place, 0, len(places))
	roomsByID := make(map[string]domain.Place, len(places))
	for _, place := range places {
		if place.Type != "classroom" {
			continue
		}

		placeCopy := place
		placeCopy.Tags = append([]string(nil), place.Tags...)
		rooms = append(rooms, placeCopy)
		roomsByID[placeCopy.ID] = placeCopy
	}

	sort.Slice(rooms, func(i, j int) bool {
		return rooms[i].Name < rooms[j].Name
	})

	return &Repository{
		rooms:      rooms,
		roomsByID:  roomsByID,
		nextNumber: 1,
	}
}

func (r *Repository) ListAvailability(_ context.Context, filter domain.RoomAvailabilityFilter) ([]domain.RoomAvailability, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	items := make([]domain.RoomAvailability, 0, len(r.rooms))
	for _, room := range r.rooms {
		if filter.Building != "" && !strings.EqualFold(room.Coordinates.Building, filter.Building) {
			continue
		}
		if filter.Capacity != nil && *filter.Capacity > 0 {
			continue
		}
		if len(filter.Equipment) > 0 && !hasAllEquipment(room.Tags, filter.Equipment) {
			continue
		}
		if r.hasBookingConflictLocked(room.ID, filter.StartsAt, filter.EndsAt) {
			continue
		}

		items = append(items, availabilityFromPlace(room, filter.StartsAt, filter.EndsAt))
	}

	return items, nil
}

func (r *Repository) GetSchedule(_ context.Context, roomID string, date time.Time) (*domain.RoomScheduleResponse, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	room, ok := r.roomsByID[roomID]
	if !ok {
		return nil, domain.ErrNotFound
	}

	dayStart, dayEnd := dayBounds(date)
	items := make([]domain.RoomScheduleItem, 0, len(r.bookings))
	for _, booking := range r.bookings {
		request := booking.request
		if request.RoomID != roomID || !domain.TimeRangesOverlap(request.StartsAt, request.EndsAt, dayStart, dayEnd) {
			continue
		}

		items = append(items, domain.RoomScheduleItem{
			ID:            booking.id,
			Kind:          "booking",
			Title:         "Бронь аудитории",
			StartsAt:      request.StartsAt.Format(time.RFC3339),
			EndsAt:        request.EndsAt.Format(time.RFC3339),
			StartTime:     request.StartsAt.Format("15:04"),
			EndTime:       request.EndsAt.Format("15:04"),
			BookerName:    request.BookerName,
			BookerContact: request.BookerContact,
		})
	}

	sortScheduleItems(items)

	return &domain.RoomScheduleResponse{
		Room:  roomInfoFromPlace(room),
		Date:  date.Format("2006-01-02"),
		Items: items,
	}, nil
}

func (r *Repository) ListBookings(_ context.Context, date time.Time) ([]domain.RoomBooking, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	dayStart, dayEnd := dayBounds(date)
	items := make([]domain.RoomBooking, 0, len(r.bookings))
	for _, booking := range r.bookings {
		request := booking.request
		if !domain.TimeRangesOverlap(request.StartsAt, request.EndsAt, dayStart, dayEnd) {
			continue
		}

		room, ok := r.roomsByID[request.RoomID]
		if !ok {
			continue
		}

		items = append(items, domain.RoomBooking{
			ID:            booking.id,
			RoomID:        room.ID,
			RoomName:      room.Name,
			StartsAt:      request.StartsAt.Format(time.RFC3339),
			EndsAt:        request.EndsAt.Format(time.RFC3339),
			BookerName:    request.BookerName,
			BookerContact: request.BookerContact,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].StartsAt != items[j].StartsAt {
			return items[i].StartsAt < items[j].StartsAt
		}
		return items[i].RoomName < items[j].RoomName
	})

	return items, nil
}

func (r *Repository) CreateBooking(_ context.Context, request domain.RoomBookingRequest) (*domain.RoomBooking, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	room, ok := r.roomsByID[request.RoomID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	if r.hasBookingConflictLocked(request.RoomID, request.StartsAt, request.EndsAt) {
		return nil, domain.ErrConflict
	}

	id := bookingID(r.nextNumber)
	r.nextNumber++
	r.bookings = append(r.bookings, storedBooking{
		id:      id,
		request: request,
	})

	return &domain.RoomBooking{
		ID:            id,
		RoomID:        room.ID,
		RoomName:      room.Name,
		StartsAt:      request.StartsAt.Format(time.RFC3339),
		EndsAt:        request.EndsAt.Format(time.RFC3339),
		BookerName:    request.BookerName,
		BookerContact: request.BookerContact,
	}, nil
}

func (r *Repository) CancelBooking(_ context.Context, targetBookingID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for index := range r.bookings {
		if r.bookings[index].id != targetBookingID {
			continue
		}

		r.bookings = append(r.bookings[:index], r.bookings[index+1:]...)
		return nil
	}

	return domain.ErrNotFound
}

func (r *Repository) hasBookingConflictLocked(roomID string, startsAt time.Time, endsAt time.Time) bool {
	for _, booking := range r.bookings {
		request := booking.request
		if request.RoomID != roomID {
			continue
		}
		if domain.TimeRangesOverlap(request.StartsAt, request.EndsAt, startsAt, endsAt) {
			return true
		}
	}

	return false
}

func availabilityFromPlace(place domain.Place, startsAt time.Time, endsAt time.Time) domain.RoomAvailability {
	return domain.RoomAvailability{
		RoomID:        place.ID,
		Name:          place.Name,
		Building:      place.Coordinates.Building,
		Floor:         place.Coordinates.Floor,
		Capacity:      0,
		Equipment:     append([]string(nil), place.Tags...),
		AvailableFrom: startsAt.Format(time.RFC3339),
		AvailableTo:   endsAt.Format(time.RFC3339),
	}
}

func roomInfoFromPlace(place domain.Place) domain.RoomInfo {
	return domain.RoomInfo{
		RoomID:    place.ID,
		Name:      place.Name,
		Building:  place.Coordinates.Building,
		Floor:     place.Coordinates.Floor,
		Equipment: append([]string(nil), place.Tags...),
	}
}

func hasAllEquipment(actual []string, required []string) bool {
	index := make(map[string]struct{}, len(actual))
	for _, item := range actual {
		index[strings.ToLower(strings.TrimSpace(item))] = struct{}{}
	}

	for _, item := range required {
		if _, ok := index[strings.ToLower(strings.TrimSpace(item))]; !ok {
			return false
		}
	}

	return true
}

func dayBounds(date time.Time) (time.Time, time.Time) {
	year, month, day := date.Date()
	start := time.Date(year, month, day, 0, 0, 0, 0, date.Location())
	return start, start.Add(24 * time.Hour)
}

func sortScheduleItems(items []domain.RoomScheduleItem) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].StartsAt != items[j].StartsAt {
			return items[i].StartsAt < items[j].StartsAt
		}
		return items[i].Title < items[j].Title
	})
}

func bookingID(number int) string {
	return "booking-memory-" + strconv.Itoa(number)
}
