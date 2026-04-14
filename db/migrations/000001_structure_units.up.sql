CREATE TABLE IF NOT EXISTS structure_units (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    external_uuid TEXT NOT NULL UNIQUE,
    parent_external_uuid TEXT,
    code TEXT NOT NULL,
    name TEXT NOT NULL,
    node_type TEXT NOT NULL,
    course INTEGER,
    semester INTEGER,
    raw_payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_structure_units_parent_external_uuid
    ON structure_units(parent_external_uuid);

CREATE INDEX IF NOT EXISTS idx_structure_units_code
    ON structure_units(code);

CREATE INDEX IF NOT EXISTS idx_structure_units_node_type
    ON structure_units(node_type);
