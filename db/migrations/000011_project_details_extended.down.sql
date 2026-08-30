-- 000011_project_details_extended.down.sql
-- التراجع الآمن عن الأعمدة والجداول المضافة في ملف الـ up.

-- 4. حذف جدول ربط المجموعات المرئية
DROP TABLE IF EXISTS project_visibility_groups;

-- 3. حذف جدول إعدادات Grazie
DROP TABLE IF EXISTS project_grazie_settings;

-- 2. حذف جدول إعدادات HelpDesk
DROP TABLE IF EXISTS project_helpdesk_settings;

-- 1. إزالة الأعمدة المضافة إلى project_teams
ALTER TABLE project_teams
    DROP COLUMN IF EXISTS description,
    DROP COLUMN IF EXISTS icon,
    DROP COLUMN IF EXISTS audit_target_id,
    DROP COLUMN IF EXISTS all_users_group,
    DROP COLUMN IF EXISTS is_updatable,
    DROP COLUMN IF EXISTS is_removable;
