package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
)

type Repository struct {
	db                 *sql.DB
	rooms              []domain.Place
	roomsByID          map[string]domain.Place
	academicWeek1Start time.Time
}

func New(db *sql.DB, places []domain.Place, academicWeek1Start time.Time) *Repository {
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
		db:                 db,
		rooms:              rooms,
		roomsByID:          roomsByID,
		academicWeek1Start: academicWeek1Start,
	}
}

func (r *Repository) ListAvailability(ctx context.Context, filter domain.RoomAvailabilityFilter) ([]domain.RoomAvailability, error) {
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

		busyBySchedule, err := r.hasScheduleConflict(ctx, room, filter.StartsAt, filter.EndsAt)
		if err != nil {
			return nil, err
		}
		if busyBySchedule {
			continue
		}

		busyByBooking, err := r.hasBookingConflict(ctx, room.ID, filter.StartsAt, filter.EndsAt)
		if err != nil {
			return nil, err
		}
		if busyByBooking {
			continue
		}

		items = append(items, availabilityFromPlace(room, filter.StartsAt, filter.EndsAt))
	}

	return items, nil
}

func (r *Repository) GetSchedule(ctx context.Context, roomID string, date time.Time) (*domain.RoomScheduleResponse, error) {
	room, ok := r.roomsByID[roomID]
	if !ok {
		var found bool
		room, found = domain.SyntheticClassroomFromPlaceID(roomID)
		if !found {
			return nil, domain.ErrNotFound
		}
	}

	items, err := r.loadScheduleEvents(ctx, room, date)
	if err != nil {
		return nil, err
	}

	bookings, err := r.loadBookings(ctx, roomID, date)
	if err != nil {
		return nil, err
	}
	items = append(items, bookings...)
	sortScheduleItems(items)

	return &domain.RoomScheduleResponse{
		Room:  roomInfoFromPlace(room),
		Date:  date.Format("2006-01-02"),
		Items: items,
	}, nil
}

func (r *Repository) ListBookings(ctx context.Context, date time.Time) ([]domain.RoomBooking, error) {
	dayStart, dayEnd := dayBounds(date)
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			id::text,
			place_id,
			room_name,
			starts_at,
			ends_at,
			booker_name,
			booker_contact
		FROM room_bookings
		WHERE starts_at < $2
		  AND ends_at > $1
		ORDER BY starts_at, room_name
	`, dayStart, dayEnd)
	if err != nil {
		return nil, fmt.Errorf("query room bookings: %w", err)
	}
	defer rows.Close()

	items := make([]domain.RoomBooking, 0, 16)
	for rows.Next() {
		var (
			id            string
			roomID        string
			roomName      string
			startsAt      time.Time
			endsAt        time.Time
			bookerName    string
			bookerContact string
		)
		if err := rows.Scan(&id, &roomID, &roomName, &startsAt, &endsAt, &bookerName, &bookerContact); err != nil {
			return nil, fmt.Errorf("scan room booking: %w", err)
		}

		items = append(items, domain.RoomBooking{
			ID:            id,
			RoomID:        roomID,
			RoomName:      roomName,
			StartsAt:      startsAt.Format(time.RFC3339),
			EndsAt:        endsAt.Format(time.RFC3339),
			BookerName:    bookerName,
			BookerContact: bookerContact,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate room bookings: %w", err)
	}

	return items, nil
}

func (r *Repository) CreateBooking(ctx context.Context, request domain.RoomBookingRequest) (*domain.RoomBooking, error) {
	room, ok := r.roomsByID[request.RoomID]
	if !ok {
		return nil, domain.ErrNotFound
	}

	busyBySchedule, err := r.hasScheduleConflict(ctx, room, request.StartsAt, request.EndsAt)
	if err != nil {
		return nil, err
	}
	if busyBySchedule {
		return nil, domain.ErrConflict
	}

	var (
		id            string
		startsAt      time.Time
		endsAt        time.Time
		bookerName    string
		bookerContact string
	)
	err = r.db.QueryRowContext(ctx, `
		INSERT INTO room_bookings (
			place_id,
			room_name,
			starts_at,
			ends_at,
			booker_name,
			booker_contact
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id::text, starts_at, ends_at, booker_name, booker_contact
	`,
		room.ID,
		room.Name,
		request.StartsAt,
		request.EndsAt,
		request.BookerName,
		request.BookerContact,
	).Scan(&id, &startsAt, &endsAt, &bookerName, &bookerContact)
	if err != nil {
		if isPgConflict(err) {
			return nil, domain.ErrConflict
		}

		return nil, fmt.Errorf("insert room booking: %w", err)
	}

	return &domain.RoomBooking{
		ID:            id,
		RoomID:        room.ID,
		RoomName:      room.Name,
		StartsAt:      startsAt.Format(time.RFC3339),
		EndsAt:        endsAt.Format(time.RFC3339),
		BookerName:    bookerName,
		BookerContact: bookerContact,
	}, nil
}

func (r *Repository) CancelBooking(ctx context.Context, bookingID string) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM room_bookings
		WHERE id::text = $1
	`, bookingID)
	if err != nil {
		return fmt.Errorf("delete room booking: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read deleted booking count: %w", err)
	}
	if affected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *Repository) hasScheduleConflict(ctx context.Context, room domain.Place, startsAt time.Time, endsAt time.Time) (bool, error) {
	match := matchForPlace(room)
	if match.normalizedRoom == "" && match.externalUUID == "" {
		return false, nil
	}

	localStart := startsAt.In(time.Local)
	localEnd := endsAt.In(time.Local)
	var hasConflict bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM schedule_events se
			JOIN schedule_event_audiences sea ON sea.event_id = se.id
			WHERE se.weekday = $1
			  AND se.week_type IN ('all', $2)
			  AND se.starts_at < $4::time
			  AND se.ends_at > $3::time
			  AND (
				($5 <> '' AND sea.external_uuid = $5)
				OR (
					$6 <> ''
					AND sea.normalized_room = $6
					AND ($7 = '' OR sea.normalized_building = $7)
				)
			  )
		)
	`,
		academicWeekday(localStart),
		domain.AcademicWeekType(localStart, r.academicWeek1Start),
		clockValue(localStart),
		clockValue(localEnd),
		match.externalUUID,
		match.normalizedRoom,
		match.normalizedBuilding,
	).Scan(&hasConflict)
	if err != nil {
		return false, fmt.Errorf("query schedule conflict: %w", err)
	}

	return hasConflict, nil
}

func (r *Repository) hasBookingConflict(ctx context.Context, roomID string, startsAt time.Time, endsAt time.Time) (bool, error) {
	var hasConflict bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM room_bookings
			WHERE place_id = $1
			  AND starts_at < $3
			  AND ends_at > $2
		)
	`, roomID, startsAt, endsAt).Scan(&hasConflict)
	if err != nil {
		return false, fmt.Errorf("query booking conflict: %w", err)
	}

	return hasConflict, nil
}

func (r *Repository) loadScheduleEvents(ctx context.Context, room domain.Place, date time.Time) ([]domain.RoomScheduleItem, error) {
	match := matchForPlace(room)
	if match.normalizedRoom == "" && match.externalUUID == "" {
		return nil, nil
	}

	weekType := domain.AcademicWeekType(date, r.academicWeek1Start)
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			se.id::text,
			se.week_type,
			se.starts_at_hour,
			se.starts_at_minute,
			se.ends_at_hour,
			se.ends_at_minute,
			coalesce(
				nullif(d.short_name, ''),
				nullif(d.full_name, ''),
				nullif(d.abbr, ''),
				nullif(se.stream_name, ''),
				'Занятие'
			) AS title,
			coalesce(se.metadata, '{}'::jsonb),
			coalesce(string_agg(DISTINCT nullif(t.full_name, ''), ', '), '') AS teachers
		FROM schedule_events se
		JOIN schedule_event_audiences sea ON sea.event_id = se.id
		LEFT JOIN disciplines d ON d.id = se.discipline_id
		LEFT JOIN schedule_event_teachers setg ON setg.event_id = se.id
		LEFT JOIN teachers t ON t.id = setg.teacher_id
		WHERE se.weekday = $1
		  AND se.week_type IN ('all', $2)
		  AND (
			($3 <> '' AND sea.external_uuid = $3)
			OR (
				$4 <> ''
				AND sea.normalized_room = $4
				AND ($5 = '' OR sea.normalized_building = $5)
			)
		  )
		GROUP BY
			se.id,
			se.week_type,
			se.starts_at_hour,
			se.starts_at_minute,
			se.ends_at_hour,
			se.ends_at_minute,
			d.short_name,
			d.full_name,
			d.abbr,
			se.stream_name,
			se.metadata
		ORDER BY se.starts_at_hour, se.starts_at_minute, title
	`,
		academicWeekday(date),
		weekType,
		match.externalUUID,
		match.normalizedRoom,
		match.normalizedBuilding,
	)
	if err != nil {
		return nil, fmt.Errorf("query room schedule: %w", err)
	}
	defer rows.Close()

	items := make([]domain.RoomScheduleItem, 0, 8)
	for rows.Next() {
		var (
			id         string
			itemWeek   string
			startHour  int
			startMin   int
			endHour    int
			endMin     int
			title      string
			metadata   []byte
			teacherCSV string
		)
		if err := rows.Scan(&id, &itemWeek, &startHour, &startMin, &endHour, &endMin, &title, &metadata, &teacherCSV); err != nil {
			return nil, fmt.Errorf("scan room schedule: %w", err)
		}

		startsAt := combineDateTime(date, startHour, startMin)
		endsAt := combineDateTime(date, endHour, endMin)
		items = append(items, domain.RoomScheduleItem{
			ID:        id,
			Kind:      "lesson",
			Title:     title,
			StartsAt:  startsAt.Format(time.RFC3339),
			EndsAt:    endsAt.Format(time.RFC3339),
			StartTime: startsAt.Format("15:04"),
			EndTime:   endsAt.Format("15:04"),
			Week:      itemWeek,
			Groups:    groupsFromMetadata(metadata),
			Teachers:  splitCSV(teacherCSV),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate room schedule: %w", err)
	}

	return items, nil
}

func (r *Repository) loadBookings(ctx context.Context, roomID string, date time.Time) ([]domain.RoomScheduleItem, error) {
	dayStart, dayEnd := dayBounds(date)
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			id::text,
			starts_at,
			ends_at,
			booker_name,
			booker_contact
		FROM room_bookings
		WHERE place_id = $1
		  AND starts_at < $3
		  AND ends_at > $2
		ORDER BY starts_at
	`, roomID, dayStart, dayEnd)
	if err != nil {
		return nil, fmt.Errorf("query room bookings: %w", err)
	}
	defer rows.Close()

	items := make([]domain.RoomScheduleItem, 0, 4)
	for rows.Next() {
		var (
			id            string
			startsAt      time.Time
			endsAt        time.Time
			bookerName    string
			bookerContact string
		)
		if err := rows.Scan(&id, &startsAt, &endsAt, &bookerName, &bookerContact); err != nil {
			return nil, fmt.Errorf("scan room booking: %w", err)
		}

		items = append(items, domain.RoomScheduleItem{
			ID:            id,
			Kind:          "booking",
			Title:         "Бронь аудитории",
			StartsAt:      startsAt.Format(time.RFC3339),
			EndsAt:        endsAt.Format(time.RFC3339),
			StartTime:     startsAt.In(time.Local).Format("15:04"),
			EndTime:       endsAt.In(time.Local).Format("15:04"),
			BookerName:    bookerName,
			BookerContact: bookerContact,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate room bookings: %w", err)
	}

	return items, nil
}

type roomMatch struct {
	externalUUID       string
	normalizedRoom     string
	normalizedBuilding string
}

func matchForPlace(place domain.Place) roomMatch {
	return roomMatch{
		externalUUID:       strings.TrimSpace(place.ExternalUUID),
		normalizedRoom:     domain.NormalizeRoomName(place.Name),
		normalizedBuilding: domain.NormalizeBuildingName(place.Coordinates.Building),
	}
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

func groupsFromMetadata(raw []byte) []string {
	type eventMetadata struct {
		Groups []domain.ScheduleGroupRef `json:"groups"`
	}

	var metadata eventMetadata
	if len(raw) == 0 || json.Unmarshal(raw, &metadata) != nil {
		return nil
	}

	groups := make([]string, 0, len(metadata.Groups))
	seen := make(map[string]struct{}, len(metadata.Groups))
	for _, group := range metadata.Groups {
		name := strings.TrimSpace(group.Name)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		groups = append(groups, name)
		seen[name] = struct{}{}
	}

	return groups
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			items = append(items, part)
		}
	}

	return items
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

func academicWeekday(date time.Time) int {
	switch date.Weekday() {
	case time.Sunday:
		return 7
	default:
		return int(date.Weekday())
	}
}

func clockValue(value time.Time) string {
	return value.Format("15:04")
}

func combineDateTime(date time.Time, hour int, minute int) time.Time {
	year, month, day := date.Date()
	return time.Date(year, month, day, hour, minute, 0, 0, date.Location())
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

func isPgConflict(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}

	return pgErr.Code == "23P01" || pgErr.Code == "23505"
}
