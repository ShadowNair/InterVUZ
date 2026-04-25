package postgres

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
)

type Repository struct {
	db                 *sql.DB
	targetRootExternal string
}

func New(db *sql.DB, targetRootExternal string) *Repository {
	return &Repository{db: db, targetRootExternal: strings.TrimSpace(targetRootExternal)}
}

func (r *Repository) Import(_ context.Context, request domain.ScheduleImportRequest) (*domain.ScheduleImportResult, error) {
	return &domain.ScheduleImportResult{
		ImportID:       "import-db-001",
		Status:         "queued",
		ImportedEvents: len(request.Events),
	}, nil
}

func (r *Repository) GetGroupCatalog(ctx context.Context) (*domain.GroupCatalogResponse, error) {
	query := `
		WITH RECURSIVE valid_groups AS (
			SELECT DISTINCT ag.external_uuid
			FROM academic_groups ag
			WHERE ag.external_uuid IS NOT NULL
			  AND EXISTS (
				SELECT 1
				FROM schedule_sources ss
				JOIN schedule_source_events sse ON sse.source_id = ss.id
				WHERE ss.source_type = 'group_sync'
				  AND ss.external_reference = ag.external_uuid
			  )
		), tree AS (
			SELECT
				su.external_uuid,
				su.parent_external_uuid,
				su.code,
				su.name,
				su.node_type,
				su.course,
				su.semester
			FROM structure_units su
			JOIN valid_groups vg ON vg.external_uuid = su.external_uuid
			UNION
			SELECT
				parent.external_uuid,
				parent.parent_external_uuid,
				parent.code,
				parent.name,
				parent.node_type,
				parent.course,
				parent.semester
			FROM structure_units parent
			JOIN tree child ON child.parent_external_uuid = parent.external_uuid
		)
		SELECT DISTINCT
			external_uuid,
			parent_external_uuid,
			code,
			name,
			node_type,
			course,
			semester
		FROM tree
	`
	args := []any{}
	if r.targetRootExternal != "" {
		query = `
			WITH RECURSIVE descendants AS (
				SELECT external_uuid
				FROM structure_units
				WHERE external_uuid = $1
				UNION ALL
				SELECT su.external_uuid
				FROM structure_units su
				JOIN descendants d ON su.parent_external_uuid = d.external_uuid
			), valid_groups AS (
				SELECT DISTINCT ag.external_uuid
				FROM academic_groups ag
				JOIN descendants d ON d.external_uuid = ag.external_uuid
				WHERE ag.external_uuid IS NOT NULL
				  AND EXISTS (
					SELECT 1
					FROM schedule_sources ss
					JOIN schedule_source_events sse ON sse.source_id = ss.id
					WHERE ss.source_type = 'group_sync'
					  AND ss.external_reference = ag.external_uuid
				  )
			), tree AS (
				SELECT
					su.external_uuid,
					su.parent_external_uuid,
					su.code,
					su.name,
					su.node_type,
					su.course,
					su.semester
				FROM structure_units su
				JOIN valid_groups vg ON vg.external_uuid = su.external_uuid
				UNION
				SELECT
					parent.external_uuid,
					parent.parent_external_uuid,
					parent.code,
					parent.name,
					parent.node_type,
					parent.course,
					parent.semester
				FROM structure_units parent
				JOIN tree child ON child.parent_external_uuid = parent.external_uuid
			)
			SELECT DISTINCT
				external_uuid,
				parent_external_uuid,
				code,
				name,
				node_type,
				course,
				semester
			FROM tree
		`
		args = append(args, r.targetRootExternal)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query group catalog: %w", err)
	}
	defer rows.Close()

	type dbNode struct {
		ExternalUUID     string
		ParentExternalID sql.NullString
		Code             string
		Name             string
		NodeType         string
		Course           sql.NullInt64
		Semester         sql.NullInt64
	}

	nodes := make(map[string]*domain.GroupCatalogNode)
	childrenByParent := make(map[string][]*domain.GroupCatalogNode)
	rootCandidates := make([]*domain.GroupCatalogNode, 0, 4)

	for rows.Next() {
		var row dbNode
		if err := rows.Scan(
			&row.ExternalUUID,
			&row.ParentExternalID,
			&row.Code,
			&row.Name,
			&row.NodeType,
			&row.Course,
			&row.Semester,
		); err != nil {
			return nil, fmt.Errorf("scan group catalog row: %w", err)
		}

		node := &domain.GroupCatalogNode{
			Abbr:     row.Code,
			Name:     row.Name,
			UUID:     row.ExternalUUID,
			NodeType: row.NodeType,
		}
		if row.ParentExternalID.Valid {
			node.ParentUUID = row.ParentExternalID.String
		}
		if row.Course.Valid {
			node.Course = int(row.Course.Int64)
		}
		if row.Semester.Valid {
			node.Semester = int(row.Semester.Int64)
		}

		nodes[row.ExternalUUID] = node
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate group catalog rows: %w", err)
	}
	if len(nodes) == 0 {
		return &domain.GroupCatalogResponse{Data: domain.GroupCatalogNode{}}, nil
	}

	for _, node := range nodes {
		parentUUID := strings.TrimSpace(node.ParentUUID)
		if parentUUID == "" {
			rootCandidates = append(rootCandidates, node)
			continue
		}

		parent, ok := nodes[parentUUID]
		if !ok {
			rootCandidates = append(rootCandidates, node)
			continue
		}

		childrenByParent[parent.UUID] = append(childrenByParent[parent.UUID], node)
	}

	for parentUUID, children := range childrenByParent {
		sort.Slice(children, func(i, j int) bool {
			return catalogNodeLess(*children[i], *children[j])
		})
		nodes[parentUUID].Children = make([]domain.GroupCatalogNode, 0, len(children))
		for _, child := range children {
			nodes[parentUUID].Children = append(nodes[parentUUID].Children, *child)
		}
	}

	sort.Slice(rootCandidates, func(i, j int) bool {
		return catalogNodeLess(*rootCandidates[i], *rootCandidates[j])
	})

	return &domain.GroupCatalogResponse{Data: *rootCandidates[0]}, nil
}

func (r *Repository) GetGroupSchedule(ctx context.Context, groupID string) (*domain.GroupScheduleResponse, error) {
	query := `
		SELECT ss.raw_payload
		FROM schedule_sources ss
		WHERE ss.source_type = 'group_sync'
		  AND ss.external_reference = $1
		  AND EXISTS (
			SELECT 1
			FROM schedule_source_events sse
			WHERE sse.source_id = ss.id
		  )
	`
	args := []any{groupID}
	if r.targetRootExternal != "" {
		query += `
		  AND EXISTS (
			WITH RECURSIVE descendants AS (
				SELECT external_uuid
				FROM structure_units
				WHERE external_uuid = $2
				UNION ALL
				SELECT su.external_uuid
				FROM structure_units su
				JOIN descendants d ON su.parent_external_uuid = d.external_uuid
			)
			SELECT 1 FROM descendants d WHERE d.external_uuid = $1
		  )
		`
		args = append(args, r.targetRootExternal)
	}

	var rawPayload []byte
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&rawPayload)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("load raw group schedule: %w", err)
	}

	var response domain.GroupScheduleResponse
	if err := json.Unmarshal(rawPayload, &response); err != nil {
		return nil, fmt.Errorf("decode raw group schedule: %w", err)
	}
	if strings.TrimSpace(response.Data.UUID) == "" {
		response.Data.UUID = groupID
	}

	return &response, nil
}

func (r *Repository) GetEvent(ctx context.Context, eventID string) (*domain.ScheduleEventResponse, error) {
	const query = `
		SELECT
			se.id,
			se.external_id,
			se.weekday,
			se.lesson_number,
			se.week_type,
			se.starts_at_hour,
			se.starts_at_minute,
			se.ends_at_hour,
			se.ends_at_minute,
			to_char(se.starts_at, 'HH24:MI') AS start_time,
			to_char(se.ends_at, 'HH24:MI') AS end_time,
			coalesce(se.permission, ''),
			coalesce(se.metadata, '{}'::jsonb),
			coalesce(d.abbr, ''),
			coalesce(se.lesson_type, ''),
			coalesce(d.full_name, ''),
			coalesce(d.short_name, '')
		FROM schedule_events se
		LEFT JOIN disciplines d ON d.id = se.discipline_id
		WHERE se.id::text = $1 OR se.external_id = $1
		LIMIT 1
	`

	var (
		dbID            string
		externalID      string
		weekday         int
		lessonNumber    int
		weekType        string
		startHour       int
		startMinute     int
		endHour         int
		endMinute       int
		startTime       string
		endTime         string
		permission      string
		metadataRaw     []byte
		disciplineAbbr  string
		lessonType      string
		disciplineFull  string
		disciplineShort string
	)

	err := r.db.QueryRowContext(ctx, query, eventID).Scan(
		&dbID,
		&externalID,
		&weekday,
		&lessonNumber,
		&weekType,
		&startHour,
		&startMinute,
		&endHour,
		&endMinute,
		&startTime,
		&endTime,
		&permission,
		&metadataRaw,
		&disciplineAbbr,
		&lessonType,
		&disciplineFull,
		&disciplineShort,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("load event: %w", err)
	}

	type eventMetadata struct {
		Groups        []domain.ScheduleGroupRef `json:"groups"`
		Stream        domain.ScheduleStream     `json:"stream"`
		Audiences     []domain.Audience         `json:"audiences"`
		RawDiscipline domain.Discipline         `json:"rawDiscipline"`
	}

	var metadata eventMetadata
	if len(metadataRaw) > 0 {
		if err := json.Unmarshal(metadataRaw, &metadata); err != nil {
			return nil, fmt.Errorf("decode event metadata: %w", err)
		}
	}

	teachers, err := r.loadEventTeachers(ctx, dbID)
	if err != nil {
		return nil, err
	}

	discipline := metadata.RawDiscipline
	if strings.TrimSpace(discipline.Abbr) == "" {
		discipline.Abbr = disciplineAbbr
	}
	if strings.TrimSpace(discipline.ActType) == "" {
		discipline.ActType = lessonType
	}
	if strings.TrimSpace(discipline.FullName) == "" {
		discipline.FullName = disciplineFull
	}
	if strings.TrimSpace(discipline.ShortName) == "" {
		discipline.ShortName = disciplineShort
	}

	response := &domain.ScheduleEventResponse{
		Data: domain.ScheduleEventItem{
			Day:              weekday,
			Time:             lessonNumber,
			Week:             weekType,
			Groups:           metadata.Groups,
			Stream:           metadata.Stream,
			EndTime:          endTime,
			Teachers:         teachers,
			Audiences:        metadata.Audiences,
			StartTime:        startTime,
			Discipline:       discipline,
			Permission:       permission,
			EndTimeMinNum:    endMinute,
			EndTimeHourNum:   endHour,
			StartTimeMinNum:  startMinute,
			StartTimeHourNum: startHour,
		},
	}

	_ = externalID
	return response, nil
}

func (r *Repository) loadEventTeachers(ctx context.Context, eventID string) ([]domain.Teacher, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			coalesce(t.external_uuid, ''),
			coalesce(t.last_name, ''),
			coalesce(t.first_name, ''),
			coalesce(t.middle_name, '')
		FROM schedule_event_teachers setg
		JOIN teachers t ON t.id = setg.teacher_id
		WHERE setg.event_id = $1
		ORDER BY t.full_name
	`, eventID)
	if err != nil {
		return nil, fmt.Errorf("query event teachers: %w", err)
	}
	defer rows.Close()

	teachers := make([]domain.Teacher, 0, 4)
	for rows.Next() {
		var teacher domain.Teacher
		if err := rows.Scan(&teacher.UUID, &teacher.LastName, &teacher.FirstName, &teacher.MiddleName); err != nil {
			return nil, fmt.Errorf("scan event teacher: %w", err)
		}
		teachers = append(teachers, teacher)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate event teachers: %w", err)
	}

	return teachers, nil
}

func (r *Repository) ListGroupExternalUUIDs(ctx context.Context) ([]string, error) {
	query := `
		SELECT external_uuid
		FROM academic_groups
		WHERE external_uuid IS NOT NULL
		ORDER BY code
	`
	args := []any{}
	if r.targetRootExternal != "" {
		query = `
			WITH RECURSIVE descendants AS (
				SELECT external_uuid
				FROM structure_units
				WHERE external_uuid = $1
				UNION ALL
				SELECT su.external_uuid
				FROM structure_units su
				JOIN descendants d ON su.parent_external_uuid = d.external_uuid
			)
			SELECT ag.external_uuid
			FROM academic_groups ag
			JOIN descendants d ON d.external_uuid = ag.external_uuid
			WHERE ag.external_uuid IS NOT NULL
			ORDER BY ag.code
		`
		args = append(args, r.targetRootExternal)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query group uuids: %w", err)
	}
	defer rows.Close()

	result := make([]string, 0, 256)
	for rows.Next() {
		var externalUUID string
		if err := rows.Scan(&externalUUID); err != nil {
			return nil, fmt.Errorf("scan group uuid: %w", err)
		}
		result = append(result, externalUUID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate group uuids: %w", err)
	}

	return result, nil
}

func (r *Repository) UpsertGroupSchedule(ctx context.Context, ownerGroupExternalUUID string, payload *domain.GroupScheduleResponse) (saved int, err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	sourceID, err := upsertScheduleSource(ctx, tx, ownerGroupExternalUUID, payload)
	if err != nil {
		return 0, err
	}

	if _, err = tx.ExecContext(ctx, `DELETE FROM schedule_source_events WHERE source_id = $1`, sourceID); err != nil {
		return 0, fmt.Errorf("clear source mappings: %w", err)
	}

	for _, event := range payload.Data.Schedule {
		disciplineID, err := upsertDiscipline(ctx, tx, event.Discipline)
		if err != nil {
			return saved, err
		}

		eventID, err := upsertScheduleEvent(ctx, tx, sourceID, ownerGroupExternalUUID, event, disciplineID)
		if err != nil {
			return saved, err
		}

		if err := replaceEventGroupLinks(ctx, tx, eventID, event); err != nil {
			return saved, err
		}
		if err := replaceEventTeacherLinks(ctx, tx, eventID, event.Teachers); err != nil {
			return saved, err
		}
		if err := replaceEventAudienceLinks(ctx, tx, eventID, event.Audiences); err != nil {
			return saved, err
		}
		if _, err = tx.ExecContext(ctx, `
			INSERT INTO schedule_source_events (source_id, event_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, sourceID, eventID); err != nil {
			return saved, fmt.Errorf("link source to event: %w", err)
		}

		saved++
	}

	if _, err = tx.ExecContext(ctx, `
		DELETE FROM schedule_events se
		WHERE se.source_id = $1
		  AND NOT EXISTS (
				SELECT 1
				FROM schedule_source_events sse
				WHERE sse.event_id = se.id
		  )
	`, sourceID); err != nil {
		return saved, fmt.Errorf("cleanup orphan events: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return saved, fmt.Errorf("commit tx: %w", err)
	}

	return saved, nil
}

func upsertScheduleSource(ctx context.Context, tx *sql.Tx, ownerGroupExternalUUID string, payload *domain.GroupScheduleResponse) (string, error) {
	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal source payload: %w", err)
	}

	var sourceID string
	err = tx.QueryRowContext(ctx, `
		INSERT INTO schedule_sources (
			source_type,
			source_url,
			file_name,
			external_reference,
			imported_at,
			status,
			raw_payload
		)
		VALUES ('group_sync', $1, $2, $3, NOW(), 'completed', $4)
		ON CONFLICT (source_type, external_reference) DO UPDATE SET
			source_url = EXCLUDED.source_url,
			file_name = EXCLUDED.file_name,
			imported_at = NOW(),
			status = EXCLUDED.status,
			raw_payload = EXCLUDED.raw_payload
		RETURNING id
	`, payload.Data.Link, payload.Data.Title, ownerGroupExternalUUID, rawPayload).Scan(&sourceID)
	if err != nil {
		return "", fmt.Errorf("upsert schedule source: %w", err)
	}

	return sourceID, nil
}

func upsertDiscipline(ctx context.Context, tx *sql.Tx, discipline domain.Discipline) (*string, error) {
	if strings.TrimSpace(discipline.FullName) == "" && strings.TrimSpace(discipline.ShortName) == "" && strings.TrimSpace(discipline.Abbr) == "" {
		return nil, nil
	}

	externalID := disciplineKey(discipline)
	fullName := firstNonEmpty(discipline.FullName, discipline.ShortName, discipline.Abbr)
	shortName := nullableTrimmedString(discipline.ShortName)
	abbr := nullableTrimmedString(discipline.Abbr)

	var id string
	err := tx.QueryRowContext(ctx, `
		INSERT INTO disciplines (external_id, abbr, short_name, full_name)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (external_id) DO UPDATE SET
			abbr = EXCLUDED.abbr,
			short_name = EXCLUDED.short_name,
			full_name = EXCLUDED.full_name
		RETURNING id
	`, externalID, abbr, shortName, fullName).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("upsert discipline: %w", err)
	}

	return &id, nil
}

func upsertScheduleEvent(ctx context.Context, tx *sql.Tx, sourceID string, ownerGroupExternalUUID string, event domain.ScheduleEventItem, disciplineID *string) (string, error) {
	externalID := eventSignature(event)

	metadataPayload := map[string]any{
		"ownerGroupExternalUUID": ownerGroupExternalUUID,
		"groups":                 event.Groups,
		"stream":                 event.Stream,
		"audiences":              event.Audiences,
		"rawDiscipline":          event.Discipline,
	}
	metadataJSON, err := json.Marshal(metadataPayload)
	if err != nil {
		return "", fmt.Errorf("marshal event metadata: %w", err)
	}

	var eventID string
	err = tx.QueryRowContext(ctx, `
		INSERT INTO schedule_events (
			source_id,
			external_id,
			discipline_id,
			audience_place_id,
			stream_name,
			lesson_type,
			weekday,
			lesson_number,
			week_type,
			starts_at,
			ends_at,
			starts_at_hour,
			starts_at_minute,
			ends_at_hour,
			ends_at_minute,
			permission,
			metadata
		)
		VALUES (
			$1, $2, $3, NULL, $4, $5, $6, $7, $8,
			$9::time, $10::time, $11, $12, $13, $14, $15, $16::jsonb
		)
		ON CONFLICT (external_id) DO UPDATE SET
			source_id = EXCLUDED.source_id,
			discipline_id = EXCLUDED.discipline_id,
			stream_name = EXCLUDED.stream_name,
			lesson_type = EXCLUDED.lesson_type,
			weekday = EXCLUDED.weekday,
			lesson_number = EXCLUDED.lesson_number,
			week_type = EXCLUDED.week_type,
			starts_at = EXCLUDED.starts_at,
			ends_at = EXCLUDED.ends_at,
			starts_at_hour = EXCLUDED.starts_at_hour,
			starts_at_minute = EXCLUDED.starts_at_minute,
			ends_at_hour = EXCLUDED.ends_at_hour,
			ends_at_minute = EXCLUDED.ends_at_minute,
			permission = EXCLUDED.permission,
			metadata = EXCLUDED.metadata
		RETURNING id
	`,
		sourceID,
		externalID,
		nullableString(disciplineID),
		nullableTrimmedString(event.Stream.Name),
		nullableTrimmedString(event.Discipline.ActType),
		event.Day,
		event.Time,
		normalizeWeekType(event.Week),
		event.StartTime,
		event.EndTime,
		event.StartTimeHourNum,
		event.StartTimeMinNum,
		event.EndTimeHourNum,
		event.EndTimeMinNum,
		nullableTrimmedString(event.Permission),
		string(metadataJSON),
	).Scan(&eventID)
	if err != nil {
		return "", fmt.Errorf("upsert schedule event: %w", err)
	}

	return eventID, nil
}

func replaceEventGroupLinks(ctx context.Context, tx *sql.Tx, eventID string, event domain.ScheduleEventItem) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM schedule_event_groups WHERE event_id = $1`, eventID); err != nil {
		return fmt.Errorf("clear event groups: %w", err)
	}

	groupFlags := make(map[string][2]int, len(event.Stream.Groups))
	for _, item := range event.Stream.Groups {
		groupFlags[item.GroupUUID] = [2]int{item.Sub1, item.Sub2}
	}

	externalUUIDs := make([]string, 0, len(groupFlags)+len(event.Groups))
	seen := make(map[string]struct{})
	for _, item := range event.Groups {
		if item.UUID == "" {
			continue
		}
		if _, ok := seen[item.UUID]; !ok {
			externalUUIDs = append(externalUUIDs, item.UUID)
			seen[item.UUID] = struct{}{}
		}
	}
	for groupUUID := range groupFlags {
		if groupUUID == "" {
			continue
		}
		if _, ok := seen[groupUUID]; !ok {
			externalUUIDs = append(externalUUIDs, groupUUID)
			seen[groupUUID] = struct{}{}
		}
	}

	if len(externalUUIDs) == 0 {
		return nil
	}

	for _, externalUUID := range externalUUIDs {
		var groupID string
		err := tx.QueryRowContext(ctx, `SELECT id FROM academic_groups WHERE external_uuid = $1`, externalUUID).Scan(&groupID)
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		if err != nil {
			return fmt.Errorf("resolve academic group %s: %w", externalUUID, err)
		}

		subgroups := groupFlags[externalUUID]
		isPrimary := hasPrimaryGroup(event.Groups, externalUUID)
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO schedule_event_groups (event_id, group_id, is_primary, subgroup_1, subgroup_2)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (event_id, group_id) DO UPDATE SET
				is_primary = EXCLUDED.is_primary,
				subgroup_1 = EXCLUDED.subgroup_1,
				subgroup_2 = EXCLUDED.subgroup_2
		`, eventID, groupID, isPrimary, subgroups[0], subgroups[1]); err != nil {
			return fmt.Errorf("insert event group link: %w", err)
		}
	}

	return nil
}

func replaceEventTeacherLinks(ctx context.Context, tx *sql.Tx, eventID string, teachers []domain.Teacher) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM schedule_event_teachers WHERE event_id = $1`, eventID); err != nil {
		return fmt.Errorf("clear event teachers: %w", err)
	}

	for _, teacher := range teachers {
		if strings.TrimSpace(teacher.UUID) == "" && strings.TrimSpace(teacher.LastName) == "" && strings.TrimSpace(teacher.FirstName) == "" {
			continue
		}

		fullName := strings.TrimSpace(strings.Join([]string{
			strings.TrimSpace(teacher.LastName),
			strings.TrimSpace(teacher.FirstName),
			strings.TrimSpace(teacher.MiddleName),
		}, " "))
		fullName = strings.Join(strings.Fields(fullName), " ")

		var teacherID string
		err := tx.QueryRowContext(ctx, `
			INSERT INTO teachers (external_uuid, last_name, first_name, middle_name, full_name)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (external_uuid) DO UPDATE SET
				last_name = EXCLUDED.last_name,
				first_name = EXCLUDED.first_name,
				middle_name = EXCLUDED.middle_name,
				full_name = EXCLUDED.full_name
			RETURNING id
		`,
			nullableTrimmedString(teacher.UUID),
			firstNonEmpty(teacher.LastName, "-"),
			firstNonEmpty(teacher.FirstName, "-"),
			nullableTrimmedString(teacher.MiddleName),
			firstNonEmpty(fullName, teacher.UUID, "unknown"),
		).Scan(&teacherID)
		if err != nil {
			return fmt.Errorf("upsert teacher: %w", err)
		}

		if _, err := tx.ExecContext(ctx, `
			INSERT INTO schedule_event_teachers (event_id, teacher_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, eventID, teacherID); err != nil {
			return fmt.Errorf("insert event teacher link: %w", err)
		}
	}

	return nil
}

func replaceEventAudienceLinks(ctx context.Context, tx *sql.Tx, eventID string, audiences []domain.Audience) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM schedule_event_audiences WHERE event_id = $1`, eventID); err != nil {
		return fmt.Errorf("clear event audiences: %w", err)
	}

	for _, audience := range audiences {
		name := strings.TrimSpace(audience.Name)
		normalizedRoom := domain.NormalizeRoomName(name)
		if normalizedRoom == "" {
			continue
		}

		metadataJSON, err := json.Marshal(audience)
		if err != nil {
			return fmt.Errorf("marshal event audience: %w", err)
		}

		if _, err := tx.ExecContext(ctx, `
			INSERT INTO schedule_event_audiences (
				event_id,
				external_uuid,
				name,
				building,
				normalized_room,
				normalized_building,
				metadata
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb)
			ON CONFLICT (event_id, normalized_room, normalized_building) DO UPDATE SET
				external_uuid = EXCLUDED.external_uuid,
				name = EXCLUDED.name,
				building = EXCLUDED.building,
				metadata = EXCLUDED.metadata
		`,
			eventID,
			nullableTrimmedString(audience.UUID),
			name,
			nullableTrimmedString(audience.Building),
			normalizedRoom,
			domain.NormalizeBuildingName(audience.Building),
			string(metadataJSON),
		); err != nil {
			return fmt.Errorf("insert event audience link: %w", err)
		}
	}

	return nil
}

func catalogNodeLess(a, b domain.GroupCatalogNode) bool {
	weight := func(nodeType string) int {
		switch nodeType {
		case "faculty":
			return 1
		case "department":
			return 2
		case "course":
			return 3
		case "group":
			return 4
		default:
			return 5
		}
	}

	if weight(a.NodeType) != weight(b.NodeType) {
		return weight(a.NodeType) < weight(b.NodeType)
	}
	if a.Course != b.Course {
		if a.Course == 0 {
			return true
		}
		if b.Course == 0 {
			return false
		}
		return a.Course < b.Course
	}
	if a.Abbr != b.Abbr {
		return a.Abbr < b.Abbr
	}
	return a.UUID < b.UUID
}

func eventSignature(event domain.ScheduleEventItem) string {
	groupUUIDs := make([]string, 0, len(event.Stream.Groups)+len(event.Groups))
	seenGroups := make(map[string]struct{})
	for _, group := range event.Groups {
		if group.UUID == "" {
			continue
		}
		if _, ok := seenGroups[group.UUID]; ok {
			continue
		}
		groupUUIDs = append(groupUUIDs, group.UUID)
		seenGroups[group.UUID] = struct{}{}
	}
	for _, group := range event.Stream.Groups {
		if group.GroupUUID == "" {
			continue
		}
		if _, ok := seenGroups[group.GroupUUID]; ok {
			continue
		}
		groupUUIDs = append(groupUUIDs, group.GroupUUID)
		seenGroups[group.GroupUUID] = struct{}{}
	}
	sort.Strings(groupUUIDs)

	teacherUUIDs := make([]string, 0, len(event.Teachers))
	for _, teacher := range event.Teachers {
		value := teacher.UUID
		if value == "" {
			value = strings.TrimSpace(strings.Join([]string{teacher.LastName, teacher.FirstName, teacher.MiddleName}, "|"))
		}
		teacherUUIDs = append(teacherUUIDs, value)
	}
	sort.Strings(teacherUUIDs)

	audienceUUIDs := make([]string, 0, len(event.Audiences))
	for _, audience := range event.Audiences {
		value := audience.UUID
		if value == "" {
			value = audience.Name
		}
		audienceUUIDs = append(audienceUUIDs, value)
	}
	sort.Strings(audienceUUIDs)

	signature := strings.Join([]string{
		fmt.Sprintf("d=%d", event.Day),
		fmt.Sprintf("t=%d", event.Time),
		fmt.Sprintf("w=%s", normalizeWeekType(event.Week)),
		fmt.Sprintf("s=%s", strings.TrimSpace(event.StartTime)),
		fmt.Sprintf("e=%s", strings.TrimSpace(event.EndTime)),
		fmt.Sprintf("da=%s", strings.TrimSpace(event.Discipline.Abbr)),
		fmt.Sprintf("df=%s", strings.TrimSpace(event.Discipline.FullName)),
		fmt.Sprintf("ds=%s", strings.TrimSpace(event.Discipline.ShortName)),
		fmt.Sprintf("lt=%s", strings.TrimSpace(event.Discipline.ActType)),
		fmt.Sprintf("g=%s", strings.Join(groupUUIDs, ",")),
		fmt.Sprintf("tt=%s", strings.Join(teacherUUIDs, ",")),
		fmt.Sprintf("aa=%s", strings.Join(audienceUUIDs, ",")),
	}, "|")

	sum := sha1.Sum([]byte(signature))
	return "group-sync-" + hex.EncodeToString(sum[:])
}

func disciplineKey(discipline domain.Discipline) string {
	payload := strings.Join([]string{
		strings.TrimSpace(discipline.Abbr),
		strings.TrimSpace(discipline.ShortName),
		strings.TrimSpace(discipline.FullName),
	}, "|")
	sum := sha1.Sum([]byte(payload))
	return "discipline-" + hex.EncodeToString(sum[:])
}

func normalizeWeekType(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "ch", "zn", "all":
		return strings.TrimSpace(strings.ToLower(value))
	default:
		return "all"
	}
}

func hasPrimaryGroup(groups []domain.ScheduleGroupRef, externalUUID string) bool {
	for _, group := range groups {
		if group.UUID == externalUUID {
			return true
		}
	}
	return false
}

func nullableTrimmedString(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return value
}

func nullableString(value *string) any {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil
	}
	return strings.TrimSpace(*value)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}
