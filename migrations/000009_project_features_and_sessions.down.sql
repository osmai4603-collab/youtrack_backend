-- =============================================================================
--  000009_project_features_and_sessions.down.sql
--  =============================================================================
--  Purpose: Rollback project features and sessions table.
-- =============================================================================
DROP TABLE IF EXISTS sessions CASCADE;
DROP TABLE IF EXISTS project_templates CASCADE;
DROP TABLE IF EXISTS project_subsystems CASCADE;
DELETE FROM project_favorites WHERE user_id = 'admin';
DROP TABLE IF EXISTS project_favorites CASCADE;
ALTER TABLE projects DROP COLUMN IF EXISTS starting_number;
