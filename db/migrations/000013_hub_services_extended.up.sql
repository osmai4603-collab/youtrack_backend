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