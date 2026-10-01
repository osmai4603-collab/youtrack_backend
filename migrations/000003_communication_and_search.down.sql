-- =============================================================================
--  000003_communication_and_search.down.sql
--  =============================================================================
--  Purpose: Rollback inbox threads, banners, and search assist tables.
-- =============================================================================
DROP TABLE IF EXISTS search_assist_history CASCADE;
DROP TABLE IF EXISTS search_assist_saved CASCADE;
DROP TABLE IF EXISTS search_assist_queries CASCADE;
DROP TABLE IF EXISTS search_assist_recent CASCADE;
DROP TABLE IF EXISTS search_assist_suggestions CASCADE;
DROP TABLE IF EXISTS search_assist_filters CASCADE;
DROP TABLE IF EXISTS banner_user_dismissals CASCADE;
DROP TABLE IF EXISTS banners CASCADE;
DROP TABLE IF EXISTS inbox_notifications CASCADE;
DROP TABLE IF EXISTS inbox_participants CASCADE;
DROP TABLE IF EXISTS inbox_messages CASCADE;
DROP TABLE IF EXISTS inbox_threads CASCADE;
