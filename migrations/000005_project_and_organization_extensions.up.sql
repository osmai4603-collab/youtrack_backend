-- =============================================================================
--  000005_project_and_organization_extensions.up.sql
--  =============================================================================
--  Purpose: Project details extended, organizations extended, and hub services extended.
--  Source: init_database.sql sections [11], [12], [13] (lines 1412-1497).
--  Dependencies: 000001_core_schema.
-- =============================================================================

-- #############################################################################
--  [11] تفاصيل المشروع
--  المصدر: migrations/000011_project_details_extended.up.sql
-- #############################################################################

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


-- #############################################################################
--  [12] المنظمات
--  المصدر: migrations/000012_organizations_extended.up.sql
-- #############################################################################

-- 000012_organizations_extended.up.sql
-- Request #22: GET /api/admin/organizations
-- يضيف أعمدة ناقصة إلى organizations ويلقّح منظمة افتراضية مطابقة لبيانات الطلب
-- الحقيقي دون تعديل أي schema سابق (000001 إلى 000011).

-- 1. إضافة الأعمدة الناقصة
ALTER TABLE organizations
    ADD COLUMN IF NOT EXISTS description TEXT,
    ADD COLUMN IF NOT EXISTS audit_target_id VARCHAR(100);

-- 2. بذر منظمة افتراضية بالمعرّف '1-0' (كما في استجابة YouTrack الحقيقية)
INSERT INTO organizations (id, key, name, icon_url, projects_count, description, audit_target_id)
VALUES ('1-0', 'CFSksvCi5N6T06bFUtA8w', 'me', NULL, 0, NULL, NULL)
ON CONFLICT (id) DO UPDATE
    SET key = EXCLUDED.key,
        name = EXCLUDED.name;


-- #############################################################################
--  [13] خدمات Hub
--  المصدر: migrations/000013_hub_services_extended.up.sql
-- #############################################################################

-- 000013_hub_services_extended.up.sql
-- Request #24: GET /hub/api/rest/services
-- يضيف أعمدة خدمة Hub الإضافية إلى جدول services دون تعديل أي schema سابق
-- (000001 إلى 000012). الأعمدة مطلوبة بواجهات Hub الأخرى في hh.json و hh2.json
-- مثل iconUrl و userUriPattern.

ALTER TABLE services
    ADD COLUMN IF NOT EXISTS icon_url TEXT,
    ADD COLUMN IF NOT EXISTS user_uri_pattern TEXT,
    ADD COLUMN IF NOT EXISTS group_uri_pattern TEXT,
    ADD COLUMN IF NOT EXISTS audience VARCHAR(255),
    ADD COLUMN IF NOT EXISTS immutable BOOLEAN DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS client_credentials_flow_enabled BOOLEAN DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS auth_code_flow_enabled BOOLEAN DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS implicit_flow_enabled BOOLEAN DEFAULT FALSE;


