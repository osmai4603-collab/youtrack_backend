
-- Migration: create sessions and add issue workflow fields.
CREATE TABLE IF NOT EXISTS sessions (
    id VARCHAR(128) PRIMARY KEY,
    session_id VARCHAR(128) NOT NULL UNIQUE,
    data TEXT,
    user_id VARCHAR(64) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL
);

ALTER TABLE issues
    ADD COLUMN IF NOT EXISTS state        VARCHAR(32)  NOT NULL DEFAULT 'to-do',
    ADD COLUMN IF NOT EXISTS priority     VARCHAR(32)  NOT NULL DEFAULT 'normal',
    ADD COLUMN IF NOT EXISTS issue_type   VARCHAR(32)  NOT NULL DEFAULT 'task',
    ADD COLUMN IF NOT EXISTS assignee_id  VARCHAR(64)  REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS due_date     BIGINT,
    ADD COLUMN IF NOT EXISTS estimation   INT,
    ADD COLUMN IF NOT EXISTS spent_time   INT,
    ADD COLUMN IF NOT EXISTS subsystem_id VARCHAR(64),
    ADD COLUMN IF NOT EXISTS fix_versions TEXT;
