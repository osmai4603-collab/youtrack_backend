-- =============================================================================
--  000005_project_and_organization_extensions.down.sql
--  =============================================================================
--  Purpose: Rollback project, organization, and hub service extensions.
-- =============================================================================
ALTER TABLE services DROP COLUMN IF EXISTS home_url, DROP COLUMN IF EXISTS application_name;
ALTER TABLE organizations DROP COLUMN IF EXISTS description;
ALTER TABLE project_teams DROP COLUMN IF EXISTS parent_team_id;
DROP TABLE IF EXISTS project_member_roles CASCADE;
DROP TABLE IF EXISTS project_team_roles CASCADE;
DROP TABLE IF EXISTS project_custom_fields CASCADE;
