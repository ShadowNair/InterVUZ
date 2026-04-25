ALTER TABLE room_bookings
    DROP CONSTRAINT IF EXISTS room_bookings_no_overlap;

DROP INDEX IF EXISTS idx_room_bookings_place_time;
DROP TABLE IF EXISTS room_bookings;

DROP INDEX IF EXISTS idx_schedule_event_audiences_room_building;
DROP INDEX IF EXISTS idx_schedule_event_audiences_external_uuid;
DROP TABLE IF EXISTS schedule_event_audiences;
