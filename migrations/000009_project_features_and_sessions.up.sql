-- =============================================================================
--  000009_project_features_and_sessions.up.sql
--  =============================================================================
--  Purpose: Project favorites, starting_number, subsystems, templates, and sessions.
--  Source: init_database.sql sections [20], [22], [23], [24], [25], [26] (lines 2016-2048, 2222-2397).
--  Dependencies: 000001_core_schema, 000007_identity_bootstrap_and_projects_seed.
-- =============================================================================

-- #############################################################################
--  [20] المفضلة لكل هوية
--  المصدر: migrations/000020_project_favorites.up.sql
-- #############################################################################

-- 000020_project_favorites.up.sql
--
-- Per-identity project favorites, exposed as the `pinned` field of
-- GET /api/projects/me.
--
-- projects.pinned is a single global flag: it cannot express "this account
-- starred this project", and no endpoint ever wrote it, so every identity saw
-- the same value. This table holds the real per-user state. The repository now
-- derives `pinned` from here, and projects.pinned is left in place only as a
-- legacy column (it still carries the seeded defaults for 0-0 and 0-1).
--
-- The composite primary key makes "star" idempotent: repeating the same request
-- is an ON CONFLICT no-op rather than a duplicate row, and un-starring is a
-- single delete.

CREATE TABLE IF NOT EXISTS project_favorites (
    user_id    VARCHAR(20) NOT NULL REFERENCES users(id)    ON DELETE CASCADE,
    project_id VARCHAR(20) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    created_at BIGINT      NOT NULL,
    PRIMARY KEY (user_id, project_id)
);

-- Listing the identities that starred a project keys on project_id; the primary
-- key already covers the user_id -> project_id direction the list endpoint uses.
CREATE INDEX IF NOT EXISTS idx_project_favorites_project_id
    ON project_favorites (project_id);



-- #############################################################################
--  [22] بداية ترقيم المسائل
--  المصدر: migrations/000022_project_starting_number.up.sql
-- #############################################################################

-- 000022_project_starting_number.up.sql
--
-- Backing store for the per-project issue numbering start.
--
-- The frontend used to keep this number in a cubit field and round-trip it to a
-- repository method that threw UnimplementedError, so the value never left the
-- device. Persisting it here makes the setting survive a restart and lets other
-- clients read it.
--
-- Default 1 matches the previous hard-coded start, so every existing project
-- keeps numbering its issues exactly as it did.

ALTER TABLE projects
    ADD COLUMN IF NOT EXISTS starting_number BIGINT NOT NULL DEFAULT 1;


-- #############################################################################
--  [23] نقل المفضلة
--  المصدر: migrations/000023_project_favorites_seed.up.sql
-- #############################################################################

-- 000023_project_favorites_seed.up.sql
--
-- Carries the pre-existing pins over to the per-identity favorites table.
--
-- Migration 000019 seeded projects.pinned = TRUE for the demo projects, and the
-- projects screen rendered them starred. Since 000020 that global column is no
-- longer read: `pinned` is computed per identity, so those stars would silently
-- disappear for every account. This migration re-creates them for the seed
-- account (login 'admin') so the screen looks the same as before.
--
-- The account is resolved by login rather than by a hardcoded id. 000019 could
-- not use a literal '2-1' either: login is UNIQUE, and the ids that login owns
-- in a given database depend on what that database already contained, so 000019
-- settles the conflict by giving the seed user the id 'admin' and referring to
-- it that way. Naming '2-1' here made the migration fail outright on any
-- database that took that other branch, because the foreign key on user_id
-- rejects a user that does not exist:
--
--   ERROR: insert or update on table "project_favorites" violates foreign key
--   constraint "project_favorites_user_id_fkey" (SQLSTATE 23503)
--
-- and a migration that cannot apply stops the server from starting at all.
-- users.login is UNIQUE, so the join yields at most one row; when no such
-- account exists the insert simply matches nothing and the account starts with
-- no favorites, which is the correct default for a per-user flag.
--
-- The project side is derived from the column rather than naming the project
-- ids, so it follows whatever 000019 actually pinned.
--
-- projects.pinned is intentionally left untouched: it is no longer part of any
-- read path, and rewriting it would fight 000019's ON CONFLICT DO UPDATE clause
-- on every re-run.

INSERT INTO project_favorites (user_id, project_id, created_at)
SELECT
    u.id,
    p.id,
    (EXTRACT(EPOCH FROM now()) * 1000)::BIGINT
FROM projects p
CROSS JOIN users u
WHERE COALESCE(p.pinned, false) = TRUE
  AND u.login = 'admin'
ON CONFLICT (user_id, project_id) DO NOTHING;


CREATE TABLE IF NOT EXISTS project_subsystems (
    id         VARCHAR(20)  PRIMARY KEY,
    project_id VARCHAR(20)  NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name       VARCHAR(255) NOT NULL,
    owner_id   VARCHAR(20)  REFERENCES users(id) ON DELETE SET NULL,
    color      BIGINT       NOT NULL DEFAULT 0,
    created_at BIGINT       NOT NULL,
    updated_at BIGINT
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_project_subsystems_project_name
    ON project_subsystems (project_id, name);


CREATE TABLE IF NOT EXISTS project_templates (
    id             VARCHAR(20)  PRIMARY KEY,
    name           VARCHAR(255) NOT NULL,
    description    TEXT,
    icon_key       VARCHAR(100),
    default_fields JSONB        NOT NULL DEFAULT '{}'::JSONB,
    created_at     BIGINT       NOT NULL,
    updated_at     BIGINT
);

INSERT INTO project_templates (id, name, description, icon_key, default_fields, created_at)
VALUES
    ('0-0', 'Scrum',
     'Sprints, backlog and a burndown chart.',
     'scrum',
     '{"sprint": "enabled", "backlog": "enabled", "burndown": "enabled"}'::JSONB,
     (EXTRACT(EPOCH FROM now()) * 1000)::BIGINT),
    ('0-1', 'Kanban',
     'A board with columns and swimlanes.',
     'kanban',
     '{"kanban": "enabled", "swimlanes": "enabled"}'::JSONB,
     (EXTRACT(EPOCH FROM now()) * 1000)::BIGINT),
    ('0-2', 'Open Space',
     'No process enforced; issues are tracked loosely.',
     'open-space',
     '{}'::JSONB,
     (EXTRACT(EPOCH FROM now()) * 1000)::BIGINT)
ON CONFLICT (id) DO UPDATE
    SET name           = EXCLUDED.name,
        description    = EXCLUDED.description,
        icon_key       = EXCLUDED.icon_key,
        default_fields = EXCLUDED.default_fields;


CREATE TABLE IF NOT EXISTS sessions (
    id VARCHAR(128) PRIMARY KEY,
    session_id VARCHAR(128) NOT NULL UNIQUE,
    data TEXT,
    user_id VARCHAR(20) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions (user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions (expires_at);

