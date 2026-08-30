-- 000010_global_settings_allowed_origins.up.sql
-- New independent schema to store the allowedOrigins of global settings
-- (request19.txt: /api/admin/globalSettings?fields=restSettings(allowAllOrigins,allowedOrigins),...).
-- No existing table is modified.

CREATE TABLE global_settings_allowed_origins (
    id SERIAL PRIMARY KEY,
    settings_id INT REFERENCES global_settings(id) ON DELETE CASCADE,
    origin VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_global_settings_allowed_origins_settings ON global_settings_allowed_origins(settings_id);

-- Seed a default row for global_settings to guarantee reliable data retrieval.
INSERT INTO global_settings (
    allow_all_origins,
    image_text_recognition_enabled,
    ocr_supported,
    email_settings_enabled,
    email_settings_is_default,
    version,
    build,
    read_only,
    statistics_enabled,
    helpdesk_enabled
) VALUES (
    FALSE,
    TRUE,
    'true',
    TRUE,
    TRUE,
    '2025.1',
    '0',
    FALSE,
    TRUE,
    FALSE
)
ON CONFLICT DO NOTHING;
