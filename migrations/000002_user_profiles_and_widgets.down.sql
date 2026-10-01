-- =============================================================================
--  000002_user_profiles_and_widgets.down.sql
--  =============================================================================
--  Purpose: Rollback user profile extensions and dashboard widget extensions.
-- =============================================================================
ALTER TABLE user_profiles DROP COLUMN IF EXISTS date_field_pattern;
ALTER TABLE dashboard_widgets DROP COLUMN IF EXISTS config, DROP COLUMN IF EXISTS widget_type;
DROP TABLE IF EXISTS user_profile_tips_dismissed CASCADE;
DROP TABLE IF EXISTS user_profile_vcs CASCADE;
DROP TABLE IF EXISTS user_profile_appearance CASCADE;
DROP TABLE IF EXISTS user_profile_search CASCADE;
DROP TABLE IF EXISTS user_profile_notifications CASCADE;
DROP TABLE IF EXISTS user_profile_data CASCADE;
