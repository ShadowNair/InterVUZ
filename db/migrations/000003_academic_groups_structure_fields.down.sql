DROP INDEX IF EXISTS idx_academic_groups_course;
DROP INDEX IF EXISTS idx_academic_groups_faculty_external_uuid;
DROP INDEX IF EXISTS idx_academic_groups_department_external_uuid;
DROP INDEX IF EXISTS idx_academic_groups_external_uuid;

ALTER TABLE academic_groups
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS raw_payload,
    DROP COLUMN IF EXISTS course_node_name,
    DROP COLUMN IF EXISTS course_node_external_uuid,
    DROP COLUMN IF EXISTS department_code,
    DROP COLUMN IF EXISTS department_external_uuid,
    DROP COLUMN IF EXISTS faculty_code,
    DROP COLUMN IF EXISTS faculty_external_uuid;