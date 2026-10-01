-- =============================================================================
--  000008_issues_read_model_and_indexes.down.sql
--  =============================================================================
--  Purpose: Rollback issues read model and support indexes.
-- =============================================================================
DELETE FROM issue_tags WHERE issue_id IN ('0-1', '0-2', '0-3', '0-4', '0-5', '0-6');
DELETE FROM tags WHERE name IN ('bug', 'feature');
DELETE FROM issues WHERE id IN ('0-1', '0-2', '0-3', '0-4', '0-5', '0-6');
DROP INDEX IF EXISTS idx_issues_number;
DROP INDEX IF EXISTS idx_issues_state;
DROP INDEX IF EXISTS idx_issues_priority;
DROP INDEX IF EXISTS idx_issues_assignee;
ALTER TABLE issues DROP CONSTRAINT IF EXISTS chk_issues_state;
ALTER TABLE issues DROP CONSTRAINT IF EXISTS chk_issues_priority;
ALTER TABLE issues DROP CONSTRAINT IF EXISTS chk_issues_issue_type;
ALTER TABLE issues DROP CONSTRAINT IF EXISTS chk_issues_number_positive;
ALTER TABLE issues DROP CONSTRAINT IF EXISTS chk_issues_votes_nonnegative;
ALTER TABLE issues DROP COLUMN IF EXISTS number, DROP COLUMN IF EXISTS state, DROP COLUMN IF EXISTS priority, DROP COLUMN IF EXISTS issue_type, DROP COLUMN IF EXISTS summary, DROP COLUMN IF EXISTS description, DROP COLUMN IF EXISTS assignee_id, DROP COLUMN IF EXISTS resolved, DROP COLUMN IF EXISTS votes;
DROP INDEX IF EXISTS idx_issues_project_resolved;
DROP INDEX IF EXISTS idx_issues_updated;
