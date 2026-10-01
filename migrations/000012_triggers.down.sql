-- =============================================================================
--  000012_triggers.down.sql
-- =============================================================================
-- Purpose: Rollback the automatic access-control maintenance triggers.
-- =============================================================================

-- المشغّل على projects أولًا: هو من يكتب في project_teams و project_team_members
-- و assigned_roles.
DROP TRIGGER IF EXISTS trigger_after_insert_project ON projects;
DROP FUNCTION IF EXISTS assignProjectAdminRoleAfterCreateProject();

DROP TRIGGER IF EXISTS trigger_after_insert_user ON users;
DROP FUNCTION IF EXISTS createGroupMemberAfterInsertUser();

DROP TABLE IF EXISTS user_group_roles CASCADE;
