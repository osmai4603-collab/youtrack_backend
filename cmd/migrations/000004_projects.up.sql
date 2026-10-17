-- Create organizations table based on YouTrack API (Organization entity)
CREATE TABLE IF NOT EXISTS organizations (
  id           VARCHAR(64) PRIMARY KEY DEFAULT gen_random_uuid()::text,
  key          VARCHAR(64) NOT NULL UNIQUE,
  name         VARCHAR(255) NOT NULL,
  description  TEXT,
  icon_url     TEXT,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_organizations_key ON organizations (key);
CREATE INDEX IF NOT EXISTS idx_organizations_name ON organizations (name);

-- Create projects table based on YouTrack Projects API (Project entity)
CREATE TABLE IF NOT EXISTS projects (
  id               VARCHAR(64) PRIMARY KEY DEFAULT gen_random_uuid()::text,
  name             VARCHAR(255) NOT NULL,
  short_name       VARCHAR(64) NOT NULL UNIQUE,
  archived         BOOLEAN NOT NULL DEFAULT FALSE,
  template         BOOLEAN NOT NULL DEFAULT FALSE,
  description      TEXT,
  organization_id  VARCHAR(64) REFERENCES organizations (id) ON DELETE SET NULL,
  created_by       VARCHAR(64) REFERENCES users (id) ON DELETE SET NULL,
  leader_id        VARCHAR(64) REFERENCES users (id) ON DELETE SET NULL,
  from_email       VARCHAR(255),
  reply_to_email   VARCHAR(255),
  icon_url         TEXT,
  starting_number  BIGINT NOT NULL DEFAULT 1,
  team_id          VARCHAR(64) REFERENCES groups(id),
  is_public        BOOLEAN DEFAULT FALSE,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_projects_short_name ON projects (short_name);
CREATE INDEX IF NOT EXISTS idx_projects_organization_id ON projects (organization_id);

CREATE OR REPLACE FUNCTION createTeamAfterInsertProject()
RETURNS TRIGGER AS $$
DECLARE
  team_id_var VARCHAR(64);
BEGIN
  team_id_var := New.id;

  -- Create team group for the new project
  INSERT INTO groups (id, name, group_type)
  VALUES (team_id_var, NEW.short_name || ' team', 'TEAM');

  -- Create assigned role for the new team
  -- INSERT INTO assigned_roles (role_id, scope_type, project_id, group_id)
  -- VALUES ('CONTRIBUTOR', 'PROJECT', NEW.id, team_id_var);

  -- Update project's team_id
  UPDATE projects SET team_id = team_id_var WHERE id = NEW.id;

  -- If project leader is specified, add leader to team group and assign PROJECT_ADMIN role
  IF NEW.leader_id IS NOT NULL THEN
    INSERT INTO group_members (group_id, member_type, user_id, is_team_member)
    VALUES (team_id_var, 'USER', NEW.leader_id, TRUE);

    INSERT INTO assigned_roles (id, role_id, user_id, scope_type, project_id)
    VALUES (gen_random_uuid()::text, 'PROJECT_ADMIN', NEW.leader_id, 'PROJECT', NEW.id);
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS triggerAfterInsertProject ON projects;

CREATE TRIGGER triggerAfterInsertProject
  AFTER INSERT ON projects
  FOR EACH ROW EXECUTE FUNCTION createTeamAfterInsertProject();


CREATE OR REPLACE FUNCTION deleteTeamAfterDeleteProject()
RETURNS TRIGGER AS $$
BEGIN
  -- Delete the team group associated with the deleted project
  DELETE FROM groups WHERE id = OLD.team_id;

  RETURN OLD;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS triggerAfterDeleteProject ON projects;

CREATE TRIGGER triggerAfterDeleteProject
  AFTER DELETE ON projects
  FOR EACH ROW EXECUTE FUNCTION deleteTeamAfterDeleteProject();
