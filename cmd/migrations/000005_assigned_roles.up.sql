DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'access_scope_type') THEN
    CREATE TYPE access_scope_type AS ENUM ('GLOBAL', 'ORGANIZATION', 'PROJECT');
  END IF;
END$$;

CREATE TABLE IF NOT EXISTS assigned_roles (
  id              VARCHAR(64) PRIMARY KEY DEFAULT gen_random_uuid()::text,
  role_id         VARCHAR(64) NOT NULL REFERENCES roles (id) ON DELETE CASCADE,
  scope_type      access_scope_type NOT NULL,
  project_id      VARCHAR(64) REFERENCES projects(id) ON DELETE CASCADE,
  organization_id VARCHAR(64) REFERENCES organizations(id) ON DELETE CASCADE,
  user_id         VARCHAR(64) REFERENCES users (id) ON DELETE CASCADE,
  group_id        VARCHAR(64) REFERENCES groups (id) ON DELETE CASCADE,
  created_at      TIMESTAMPTZ,
  updated_at      TIMESTAMPTZ,

  -- A holder is either a user or a group, never both and never neither.
  CONSTRAINT assigned_roles_single_holder
    CHECK (num_nonnulls(user_id, group_id) = 1),

  CONSTRAINT check_scope CHECK (
    (scope_type = 'GLOBAL' AND organization_id IS NULL AND project_id IS NULL) OR
    (scope_type = 'ORGANIZATION' AND organization_id IS NOT NULL AND project_id IS NULL) OR
    (scope_type = 'PROJECT' AND organization_id IS NULL AND project_id IS NOT NULL)
  )

);

INSERT INTO assigned_roles (role_id, scope_type, group_id)
VALUES ('OBSERVER', 'GLOBAL', 'REGISTERED_USERS_GROUP')
ON CONFLICT (id) DO NOTHING;
