CREATE UNIQUE INDEX IF NOT EXISTS idx_schedule_sources_type_external_reference
    ON schedule_sources (source_type, external_reference);

CREATE TABLE IF NOT EXISTS schedule_source_events (
    source_id UUID NOT NULL REFERENCES schedule_sources(id) ON DELETE CASCADE,
    event_id UUID NOT NULL REFERENCES schedule_events(id) ON DELETE CASCADE,
    PRIMARY KEY (source_id, event_id)
);

CREATE INDEX IF NOT EXISTS idx_schedule_source_events_event_id
    ON schedule_source_events (event_id);
