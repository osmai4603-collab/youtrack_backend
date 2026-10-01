-- =============================================================================
--  000006_rbac_roles_and_auth.down.sql
--  =============================================================================
--  Purpose: Rollback RBAC seed roles and user auth columns.
-- =============================================================================
ALTER TABLE users DROP COLUMN IF EXISTS auth_type, DROP COLUMN IF EXISTS auth_data, DROP COLUMN IF EXISTS is_active, DROP COLUMN IF EXISTS last_login, DROP COLUMN IF EXISTS password_changed_at;
DELETE FROM roles WHERE id IN ('system-admin', 'project-admin', 'observer');
