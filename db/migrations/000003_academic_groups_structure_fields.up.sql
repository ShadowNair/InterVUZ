ALTER TABLE academic_groups
    ADD COLUMN IF NOT EXISTS faculty_external_uuid TEXT,
    ADD COLUMN IF NOT EXISTS faculty_code TEXT,
    ADD COLUMN IF NOT EXISTS department_external_uuid TEXT,
    ADD COLUMN IF NOT EXISTS department_code TEXT,
    ADD COLUMN IF NOT EXISTS course_node_external_uuid TEXT,
    ADD COLUMN IF NOT EXISTS course_node_name TEXT,
    ADD COLUMN IF NOT EXISTS raw_payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

CREATE INDEX IF NOT EXISTS idx_academic_groups_external_uuid
    ON academic_groups(external_uuid);

CREATE INDEX IF NOT EXISTS idx_academic_groups_department_external_uuid
    ON academic_groups(department_external_uuid);

CREATE INDEX IF NOT EXISTS idx_academic_groups_faculty_external_uuid
    ON academic_groups(faculty_external_uuid);

CREATE INDEX IF NOT EXISTS idx_academic_groups_course
    ON academic_groups(course);