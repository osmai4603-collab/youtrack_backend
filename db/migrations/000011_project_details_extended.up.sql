-- 000011_project_details_extended.up.sql
-- Request #21: GET /api/admin/projects/{id}
-- يضيف أعمدة ناقصة إلى project_teams ومجموعة جداول لإعدادات المشروع التفصيلية
-- دون تعديل أي schema سابق.

-- 1. إضافة الأعمدة الناقصة إلى project_teams
ALTER TABLE project_teams
    ADD COLUMN IF NOT EXISTS description TEXT,
    ADD COLUMN IF NOT EXISTS icon TEXT,
    ADD COLUMN IF NOT EXISTS audit_target_id VARCHAR(100),
    ADD COLUMN IF NOT EXISTS all_users_group BOOLEAN DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS is_updatable BOOLEAN DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS is_removable BOOLEAN DEFAULT FALSE;

-- 2. جدول إعدادات HelpDesk للمشاريع
CREATE TABLE IF NOT EXISTS project_helpdesk_settings (
    id VARCHAR(20) PRIMARY KEY,
    project_id VARCHAR(20) REFERENCES projects(id) ON DELETE CASCADE,
    default_form_uuid VARCHAR(100),
    default_form_title VARCHAR(255)
);

-- 3. جدول إعدادات Grazie للمشاريع
CREATE TABLE IF NOT EXISTS project_grazie_settings (
    project_id VARCHAR(20) PRIMARY KEY REFERENCES projects(id) ON DELETE CASCADE,
    disabled BOOLEAN DEFAULT FALSE
);

-- 4. جدول ربط المجموعات المرئية للمشروع
CREATE TABLE IF NOT EXISTS project_visibility_groups (
    project_id VARCHAR(20) REFERENCES projects(id) ON DELETE CASCADE,
    group_id VARCHAR(20) REFERENCES user_groups(id) ON DELETE CASCADE,
    PRIMARY KEY (project_id, group_id)
);
