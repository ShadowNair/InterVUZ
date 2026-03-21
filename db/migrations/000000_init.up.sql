CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    external_id TEXT UNIQUE,
    email TEXT UNIQUE,
    full_name TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('student', 'admin')),
    group_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE campuses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    address TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE buildings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campus_id UUID NOT NULL REFERENCES campuses(id) ON DELETE CASCADE,
    code TEXT NOT NULL,
    name TEXT NOT NULL,
    address TEXT,
    floors_count INTEGER NOT NULL DEFAULT 1 CHECK (floors_count > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (campus_id, code)
);

CREATE TABLE places (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    building_id UUID NOT NULL REFERENCES buildings(id) ON DELETE CASCADE,
    external_id TEXT UNIQUE,
    name TEXT NOT NULL,
    type TEXT NOT NULL CHECK (
        type IN (
            'classroom',
            'department',
            'cafeteria',
            'dean_office',
            'printer',
            'office',
            'library',
            'restroom',
            'entrance',
            'other'
        )
    ),
    floor INTEGER NOT NULL CHECK (floor >= 0),
    room_number TEXT,
    description TEXT,
    x NUMERIC(10,2),
    y NUMERIC(10,2),
    is_accessible BOOLEAN NOT NULL DEFAULT FALSE,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_places_building_floor ON places (building_id, floor);
CREATE INDEX idx_places_type ON places (type);

CREATE TABLE routes_cache (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_place_id UUID NOT NULL REFERENCES places(id) ON DELETE CASCADE,
    to_place_id UUID NOT NULL REFERENCES places(id) ON DELETE CASCADE,
    accessible_only BOOLEAN NOT NULL DEFAULT FALSE,
    distance_meters INTEGER NOT NULL CHECK (distance_meters >= 0),
    estimated_duration_minutes INTEGER NOT NULL CHECK (estimated_duration_minutes >= 0),
    steps JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (from_place_id, to_place_id, accessible_only)
);

CREATE TABLE academic_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    external_uuid TEXT UNIQUE,
    code TEXT NOT NULL UNIQUE,
    name TEXT,
    faculty_name TEXT,
    department_name TEXT,
    course INTEGER CHECK (course > 0),
    semester INTEGER CHECK (semester > 0),
    parent_group_id UUID REFERENCES academic_groups(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE users
    ADD CONSTRAINT fk_users_group_id
    FOREIGN KEY (group_id) REFERENCES academic_groups(id) ON DELETE SET NULL;

CREATE TABLE teachers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    external_uuid TEXT UNIQUE,
    last_name TEXT NOT NULL,
    first_name TEXT NOT NULL,
    middle_name TEXT,
    full_name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE disciplines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    external_id TEXT UNIQUE,
    abbr TEXT,
    short_name TEXT,
    full_name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE schedule_sources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    source_type TEXT NOT NULL CHECK (source_type IN ('file', 'url', 'manual', 'group_sync')),
    source_url TEXT,
    file_name TEXT,
    external_reference TEXT,
    imported_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status TEXT NOT NULL CHECK (status IN ('queued', 'processing', 'completed', 'failed')),
    raw_payload JSONB
);

CREATE TABLE schedule_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id UUID REFERENCES schedule_sources(id) ON DELETE SET NULL,
    external_id TEXT UNIQUE,
    discipline_id UUID REFERENCES disciplines(id) ON DELETE SET NULL,
    audience_place_id UUID REFERENCES places(id) ON DELETE SET NULL,
    stream_name TEXT,
    lesson_type TEXT,
    weekday SMALLINT NOT NULL CHECK (weekday BETWEEN 1 AND 7),
    lesson_number SMALLINT NOT NULL CHECK (lesson_number > 0),
    week_type TEXT NOT NULL CHECK (week_type IN ('all', 'ch', 'zn')),
    starts_at TIME NOT NULL,
    ends_at TIME NOT NULL,
    starts_at_hour SMALLINT NOT NULL CHECK (starts_at_hour BETWEEN 0 AND 23),
    starts_at_minute SMALLINT NOT NULL CHECK (starts_at_minute BETWEEN 0 AND 59),
    ends_at_hour SMALLINT NOT NULL CHECK (ends_at_hour BETWEEN 0 AND 23),
    ends_at_minute SMALLINT NOT NULL CHECK (ends_at_minute BETWEEN 0 AND 59),
    permission TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (starts_at < ends_at)
);

CREATE INDEX idx_schedule_events_slot
    ON schedule_events (weekday, lesson_number, week_type);

CREATE TABLE schedule_event_groups (
    event_id UUID NOT NULL REFERENCES schedule_events(id) ON DELETE CASCADE,
    group_id UUID NOT NULL REFERENCES academic_groups(id) ON DELETE CASCADE,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    subgroup_1 SMALLINT NOT NULL DEFAULT 0,
    subgroup_2 SMALLINT NOT NULL DEFAULT 0,
    PRIMARY KEY (event_id, group_id)
);

CREATE TABLE schedule_event_teachers (
    event_id UUID NOT NULL REFERENCES schedule_events(id) ON DELETE CASCADE,
    teacher_id UUID NOT NULL REFERENCES teachers(id) ON DELETE CASCADE,
    PRIMARY KEY (event_id, teacher_id)
);
