
CREATE TABLE IF NOT EXISTS user_group_roles (
    role_id  VARCHAR(100) REFERENCES roles(id)       ON DELETE CASCADE,
    group_id VARCHAR(20)  REFERENCES user_groups(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, group_id)
);

INSERT INTO user_group_roles (group_id, role_id)
VALUES ('registered-users', 'observer')
ON CONFLICT (group_id, role_id) DO NOTHING;

CREATE OR REPLACE FUNCTION createGroupMemberAfterInsertUser()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO user_group_members (user_id, group_id)
    VALUES (NEW.id, 'all-users'),   -- All Users
           (NEW.id, 'registered-users')     -- Registered Users
    ON CONFLICT (user_id, group_id) DO NOTHING;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_after_insert_user
AFTER INSERT ON users
FOR EACH ROW
EXECUTE FUNCTION createGroupMemberAfterInsertUser();


CREATE OR REPLACE FUNCTION assignProjectAdminRoleAfterCreateProject()
RETURNS TRIGGER AS $$
DECLARE
    
    v_assigned_role_id VARCHAR(20) := left(md5(NEW.id || ':project-admin'), 20);
    v_holder_name       VARCHAR(255);
BEGIN
    -- create a team for own project
    INSERT INTO project_teams (id, name, project_id)
    VALUES (NEW.id, NEW.short_name || ' team', NEW.id)
    ON CONFLICT (id) DO NOTHING;

    UPDATE projects SET team_id = NEW.id WHERE id = NEW.id;

    -- assign project-admin role for project owner in project scope.
    IF NEW.leader_id IS NOT NULL THEN
        SELECT login INTO v_holder_name FROM users WHERE id = NEW.leader_id;

        INSERT INTO project_team_members (team_id, user_id)
        VALUES (NEW.id, NEW.leader_id)
        ON CONFLICT (team_id, user_id) DO NOTHING;

        INSERT INTO assigned_roles (id, role_id, audit_target_id,holder_type, holder_id, holder_name,scope_type, scope_id, scope_project_id, scope_organization_id) 
        VALUES (v_assigned_role_id, 'project-admin', NULL, 'user', NEW.leader_id, v_holder_name, 'project', NEW.id, NEW.id, NULL)
        ON CONFLICT (id) DO NOTHING;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_after_insert_project AFTER INSERT ON projects FOR EACH ROW
EXECUTE FUNCTION assignProjectAdminRoleAfterCreateProject();
