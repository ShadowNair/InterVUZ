package memory

import (
	"context"
	"strings"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
)

type Repository struct {
	rooms []domain.RoomAvailability
}

func New() *Repository {
	return &Repository{
		rooms: []domain.RoomAvailability{
			{
				RoomID:        "room_405",
				Name:          "Аудитория 405",
				Building:      "B2",
				Floor:         4,
				Capacity:      30,
				Equipment:     []string{"projector", "whiteboard"},
				AvailableFrom: "2026-03-21T08:00:00Z",
				AvailableTo:   "2026-03-21T18:00:00Z",
			},
			{
				RoomID:        "room_214",
				Name:          "Аудитория 214",
				Building:      "B1",
				Floor:         2,
				Capacity:      24,
				Equipment:     []string{"projector", "computers"},
				AvailableFrom: "2026-03-21T09:00:00Z",
				AvailableTo:   "2026-03-21T17:00:00Z",
			},
		},
	}
}

func (r *Repository) ListAvailability(_ context.Context, filter domain.RoomAvailabilityFilter) ([]domain.RoomAvailability, error) {
	items := make([]domain.RoomAvailability, 0, len(r.rooms))

	for _, room := range r.rooms {
		if filter.Building != "" && !strings.EqualFold(room.Building, filter.Building) {
			continue
		}
		if filter.Capacity != nil && room.Capacity < *filter.Capacity {
			continue
		}
		if len(filter.Equipment) > 0 && !hasAllEquipment(room.Equipment, filter.Equipment) {
			continue
		}

		roomCopy := room
		roomCopy.Equipment = append([]string(nil), room.Equipment...)
		items = append(items, roomCopy)
	}

	return items, nil
}

func hasAllEquipment(actual []string, required []string) bool {
	index := make(map[string]struct{}, len(actual))
	for _, item := range actual {
		index[strings.ToLower(item)] = struct{}{}
	}

	for _, item := range required {
		if _, ok := index[strings.ToLower(item)]; !ok {
			return false
		}
	}

	return true
}
