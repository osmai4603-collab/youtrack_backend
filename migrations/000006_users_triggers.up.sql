CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE OR REPLACE FUNCTION createGroupMembersAfterInsertUser()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO group_members (group_id, member_type, user_id)
    VALUES ('ALL_USERS_GROUP', 'USER', NEW.id);
    INSERT INTO group_members (group_id, member_type, user_id)
    VALUES ('REGISTERED_USERS_GROUP', 'USER', NEW.id);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS triggerAfterInsertUser ON users;

CREATE TRIGGER triggerAfterInsertUser
  AFTER INSERT ON users
  FOR EACH ROW EXECUTE FUNCTION createGroupMembersAfterInsertUser();

INSERT INTO users (
  login, email, full_name, name, type, hash_password, created_at, updated_at
) VALUES (
  'admin', 'admin@example.com', 'System Administrator', 'admin', 'AGENT',
  crypt('admin123', gen_salt('bf', 12)), NOW(), NOW()
)
ON CONFLICT (login) DO NOTHING;

INSERT INTO assigned_roles (role_id, scope_type, user_id)
SELECT 'SYSTEM_ADMIN', 'GLOBAL', id FROM users WHERE login = 'admin' LIMIT 1;



CREATE OR REPLACE FUNCTION createAssignedRoleAfterInsertTeamMember()
RETURNS TRIGGER AS $$
DECLARE
    projectID VARCHAR(64);
BEGIN
    IF NEW.is_team_member = TRUE THEN
        SELECT id INTO projectID FROM projects WHERE team_id = NEW.group_id LIMIT 1;
        INSERT INTO assigned_roles (role_id, scope_type, project_id, user_id, group_id)
        VALUES ('CONTRIBUTOR', 'PROJECT', projectID, NEW.user_id, NEW.member_group_id);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER triggerAfterInsertTeamMember
    AFTER INSERT ON group_members
    FOR EACH ROW EXECUTE FUNCTION createAssignedRoleAfterInsertTeamMember();
