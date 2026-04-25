CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE IF NOT EXISTS schedule_event_audiences (
    event_id UUID NOT NULL REFERENCES schedule_events(id) ON DELETE CASCADE,
    external_uuid TEXT,
    name TEXT NOT NULL,
    building TEXT,
    normalized_room TEXT NOT NULL,
    normalized_building TEXT NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    PRIMARY KEY (event_id, normalized_room, normalized_building)
);

CREATE INDEX IF NOT EXISTS idx_schedule_event_audiences_external_uuid
    ON schedule_event_audiences (external_uuid)
    WHERE external_uuid IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_schedule_event_audiences_room_building
    ON schedule_event_audiences (normalized_room, normalized_building);

CREATE TABLE IF NOT EXISTS room_bookings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    place_id TEXT NOT NULL,
    room_name TEXT NOT NULL,
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL,
    booker_name TEXT NOT NULL,
    booker_contact TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (starts_at < ends_at)
);

CREATE INDEX IF NOT EXISTS idx_room_bookings_place_time
    ON room_bookings (place_id, starts_at, ends_at);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'room_bookings_no_overlap'
    ) THEN
        ALTER TABLE room_bookings
            ADD CONSTRAINT room_bookings_no_overlap
            EXCLUDE USING gist (
                place_id WITH =,
                tstzrange(starts_at, ends_at, '[)') WITH &&
            );
    END IF;
END $$;
