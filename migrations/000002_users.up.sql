DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_type') THEN
    CREATE TYPE user_type AS ENUM ('AGENT', 'STANDARD_USER', 'REPORTER');
  END IF;
END$$;

CREATE TABLE IF NOT EXISTS users (
  -- The identifier of the user.
  id               VARCHAR(64) PRIMARY KEY DEFAULT gen_random_uuid()::text,
  -- The login name of the user.
  login            VARCHAR(255) NOT NULL UNIQUE,
  -- The email address of the user.
  email            VARCHAR(255),
  -- The full name of the user (first and last name).
  full_name        VARCHAR(255),
  -- The name of the user.
  name             VARCHAR(255),
  -- The URL of the user's avatar.
  avatar_url       TEXT,
  -- Indicates whether the user is banned.
  banned           BOOLEAN NOT NULL DEFAULT FALSE,
  -- Indicates whether the user is currently online.
  online           BOOLEAN NOT NULL DEFAULT FALSE,
  -- Indicates if this user is a guest.
  guest            BOOLEAN,
  -- The type of the user.
  type             user_type,
  -- The ID of the Hub that manages this user account.
  hub_id           VARCHAR(64),
  -- The ID of the Ring that manages this user account.
  ring_id          VARCHAR(64),
  -- The timestamp of the user's last login.
  last_login_time  TIMESTAMPTZ,
  -- The timestamp of the user's creation in YouTrack.
  creation_time    TIMESTAMPTZ,
  -- The timestamp of the user's last update in YouTrack.
  update_time      TIMESTAMPTZ,
  -- Indicates whether the user's personal data has been anonymized.
  is_anonymized    BOOLEAN NOT NULL DEFAULT FALSE,
  -- The reason why the user is banned.
  ban_reason       TEXT,
  -- Hashed password of the user.
  hash_password    TEXT,
  is_locked BOOLEAN DEFAULT FALSE,
  failed_login_count INT NOT NULL DEFAULT 0,
  last_failed_login_at TIMESTAMPTZ,
  last_login_at TIMESTAMPTZ,
  locked_until TIMESTAMPTZ,
  -- Application audit timestamp when this record was created.
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  -- Application audit timestamp when this record was last updated.
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);


CREATE TABLE IF NOT EXISTS sessions (
    id VARCHAR(128) PRIMARY KEY,
    session_id VARCHAR(128) NOT NULL UNIQUE,
    data TEXT,
    user_id VARCHAR(64) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL
);
