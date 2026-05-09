CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    email      TEXT        NOT NULL UNIQUE,
    password   TEXT        NOT NULL DEFAULT '',
    role       TEXT        NOT NULL DEFAULT 'user' CHECK (role IN ('admin', 'user')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO users (id, email, role) VALUES
    ('00000000-0000-0000-0000-000000000001', 'admin@test.local', 'admin'),
    ('00000000-0000-0000-0000-000000000002', 'user@test.local',  'user')
ON CONFLICT (id) DO NOTHING;
