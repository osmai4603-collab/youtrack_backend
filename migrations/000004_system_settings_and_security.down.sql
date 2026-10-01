-- =============================================================================
--  000004_system_settings_and_security.down.sql
--  =============================================================================
--  Purpose: Rollback system settings, issue subscriptions, and security filter fields.
-- =============================================================================
DROP TABLE IF EXISTS security_filter_fields CASCADE;
DROP TABLE IF EXISTS global_settings_allowed_origins CASCADE;
DROP TABLE IF EXISTS issue_list_subscription_users CASCADE;
DROP TABLE IF EXISTS issue_list_subscriptions CASCADE;
