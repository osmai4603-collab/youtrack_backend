-- =============================================================================
--  000007_identity_bootstrap_and_projects_seed.down.sql
--  =============================================================================
--  Purpose: Rollback identity bootstrap and project seed data.
-- =============================================================================
DELETE FROM project_team_members WHERE user_id IN ('2-3', '2-4', 'admin');
DELETE FROM project_teams WHERE id IN ('2-0', '2-1', '2-2', '2-3', '2-4');
DELETE FROM projects WHERE id IN ('2-0', '2-1', '2-2', '2-3', '2-4');
DELETE FROM user_group_members WHERE user_id IN ('2-3', '2-4', 'admin');
DELETE FROM user_groups WHERE id IN ('2-0', '2-1', '2-2');
DELETE FROM assigned_roles WHERE user_id = 'admin';
DELETE FROM users WHERE id IN ('2-3', '2-4', 'admin');
DELETE FROM user_types WHERE id IN ('registered', 'guest', 'system');
DELETE FROM project_types WHERE id = 'default';
DELETE FROM organizations WHERE id = 'default';
