-- 000013_hub_services_extended.down.sql
-- التراجع الآمن عن الأعمدة المضافة في ملف الـ up.

ALTER TABLE services
    DROP COLUMN IF EXISTS implicit_flow_enabled,
    DROP COLUMN IF EXISTS auth_code_flow_enabled,
    DROP COLUMN IF EXISTS client_credentials_flow_enabled,
    DROP COLUMN IF EXISTS immutable,
    DROP COLUMN IF EXISTS audience,
    DROP COLUMN IF EXISTS group_uri_pattern,
    DROP COLUMN IF EXISTS user_uri_pattern,
    DROP COLUMN IF EXISTS icon_url;