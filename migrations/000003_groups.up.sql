CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Create groups table based on YouTrack User Groups API (UserGroup entity)
CREATE TABLE IF NOT EXISTS groups (
  -- The identifier of the user group.
  id                     VARCHAR(64) PRIMARY KEY DEFAULT gen_random_uuid()::text,
  -- The name of the group.
  name                   VARCHAR(255) NOT NULL UNIQUE,
  -- The description of the group, shown to users who can view or update it.
  description            TEXT,
  group_type             VARCHAR(10) DEFAULT 'GROUP', -- type: TEAM OR GROUP
  parent_group_id        VARCHAR(64) REFERENCES groups (id),
  -- The number of users in the group.
  users_count            BIGINT NOT NULL DEFAULT 0,
  -- The URL of the group logo.
  icon                   TEXT,
  -- True if this group contains all users, otherwise false.
  all_users_group        BOOLEAN NOT NULL DEFAULT FALSE,
  -- True if this is the default Registered Users group.
  registered_users_group BOOLEAN NOT NULL DEFAULT FALSE,
  -- True if new accounts are added to this group automatically.
  auto_join              BOOLEAN NOT NULL DEFAULT FALSE,
  -- Application audit timestamp when this record was created.
  created_at             TIMESTAMPTZ DEFAULT NOW(),
  -- Application audit timestamp when this record was last updated.
  updated_at             TIMESTAMPTZ DEFAULT NOW(),
  -- A group cannot be nested under itself.
  CONSTRAINT groups_parent_not_self CHECK (parent_group_id IS NULL OR parent_group_id <> id),
  -- The All Users and Registered Users markers are mutually exclusive.
  CONSTRAINT groups_system_markers_exclusive CHECK (NOT (all_users_group AND registered_users_group)),
  CONSTRAINT check_distinct CHECK (parent_group_id IS NULL OR parent_group_id <> id)
);

CREATE INDEX IF NOT EXISTS idx_groups_name ON groups (name);
CREATE INDEX IF NOT EXISTS idx_groups_auto_join ON groups (auto_join);

CREATE OR REPLACE FUNCTION validate_group_hierarchy()
RETURNS TRIGGER AS $$
DECLARE
    cycle_found BOOLEAN;
BEGIN

    IF NEW.parent_group_id IS NULL THEN
        RETURN NEW;
    END IF;

    WITH RECURSIVE ancestors AS (
        SELECT id, parent_group_id
        FROM groups
        WHERE id = NEW.parent_group_id

        UNION ALL

        SELECT g.id, g.parent_group_id
        FROM groups g
        JOIN ancestors a
            ON a.parent_group_id = g.id
    )
    SELECT EXISTS (
        SELECT 1
        FROM ancestors
        WHERE id = NEW.id
    )
    INTO cycle_found;

    IF cycle_found THEN
        RAISE EXCEPTION
            'Circular hierarchy detected';
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER triggerAfterInsertgGroup
    BEFORE INSERT OR UPDATE ON groups
    FOR EACH ROW EXECUTE FUNCTION validate_group_hierarchy();




-- All Users: every account is added here automatically, including the guest account.
INSERT INTO groups (id, name, all_users_group, auto_join) VALUES
    ('ALL_USERS_GROUP', 'All Users', TRUE, TRUE)
    ON CONFLICT (id) DO NOTHING;

-- Registered Users: every account created in YouTrack is added here automatically.
INSERT INTO groups (id, name, parent_group_id, all_users_group, registered_users_group, auto_join)
VALUES
    ('REGISTERED_USERS_GROUP', 'Registered Users', 'ALL_USERS_GROUP', FALSE, TRUE, TRUE)
    ON CONFLICT (id) DO NOTHING;

-- ===========================================================================
-- Group membership
-- ===========================================================================
CREATE TABLE IF NOT EXISTS group_members (
    id              VARCHAR(64) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    -- The group the user was added to directly.
    group_id        VARCHAR(64) NOT NULL REFERENCES groups (id) ON DELETE CASCADE,
    -- The user account.
    user_id         VARCHAR(64) REFERENCES users (id) ON DELETE CASCADE,
    member_type     VARCHAR(10) DEFAULT 'USER', -- type: USER OR GROUP
    member_group_id VARCHAR(64) REFERENCES groups(id) ON DELETE CASCADE,
    is_team_member  BOOLEAN DEFAULT FALSE,
    -- Application audit timestamp when this record was created.
    created_at      TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT check_members_validation CHECK (
        (member_type = 'USER'  AND user_id IS NOT NULL AND member_group_id IS NULL) OR
        (member_type = 'GROUP' AND user_id IS NULL AND member_group_id IS NOT NULL)
    ),
    CONSTRAINT chck_group CHECK (
        (member_type = 'GROUP' AND is_team_member = TRUE) OR
        (member_type = 'USER'  AND (is_team_member = TRUE OR is_team_member = FALSE))
    )
);

CREATE UNIQUE INDEX uq_group_members_user 
    ON group_members(group_id, user_id) WHERE user_id IS NOT NULL;
CREATE UNIQUE INDEX uq_group_members_group 
    ON group_members(group_id, member_group_id) WHERE member_group_id IS NOT NULL;





CREATE OR REPLACE FUNCTION updateParentGroupId()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.group_type = 'GROUP' AND NEW.parent_group_id IS NULL THEN
        UPDATE groups SET parent_group_id = 'ALL_USERS_GROUP' WHERE id = NEW.id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS triggerAfterInsertgGroup ON groups;

CREATE TRIGGER triggerAfterInsertgGroup
  AFTER INSERT ON groups
  FOR EACH ROW EXECUTE FUNCTION updateParentGroupId();


