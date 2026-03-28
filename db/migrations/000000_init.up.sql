CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    password_salt TEXT NOT NULL,
    username TEXT UNIQUE,
    name TEXT,
    surname TEXT,
    patronymic TEXT,
    avatar_url TEXT,
    role TEXT NOT NULL DEFAULT 'user',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT profiles_email_check CHECK (POSITION('@' IN email) > 1),
    CONSTRAINT profiles_password_hash_check CHECK (LENGTH(TRIM(password_hash)) > 0),
    CONSTRAINT profiles_password_salt_check CHECK (LENGTH(TRIM(password_salt)) > 0),
    CONSTRAINT profiles_role_check CHECK (role IN ('user', 'admin'))
);

CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id UUID NOT NULL REFERENCES profiles (id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ,
    CONSTRAINT refresh_tokens_expiry_check CHECK (expires_at > created_at)
);

CREATE INDEX refresh_tokens_profile_id_idx ON refresh_tokens (profile_id);
CREATE INDEX refresh_tokens_expires_at_idx ON refresh_tokens (expires_at);
