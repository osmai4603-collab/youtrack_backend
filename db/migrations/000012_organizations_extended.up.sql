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
