-- 000012_organizations_extended.down.sql
-- التراجع الآمن عن الأعمدة المضافة في ملف الـ up.

ALTER TABLE organizations
    DROP COLUMN IF EXISTS audit_target_id,
    DROP COLUMN IF EXISTS description;
