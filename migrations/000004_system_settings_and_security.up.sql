-- =============================================================================
--  000004_system_settings_and_security.up.sql
--  =============================================================================
--  Purpose: Issue list subscriptions, allowed origins, and security filter fields.
--  Source: init_database.sql sections [9], [10], [14] (lines 1328-1411, 1498-1523).
--  Dependencies: 000001_core_schema.
-- =============================================================================

-- #############################################################################
--  [9]  اشتراك قائمة المسائل
--  المصدر: migrations/000009_issue_list_subscription.up.sql
-- #############################################################################

-- 000008_issue_list_subscription.up.sql
-- New independent schema for the issue list subscription endpoint
-- (request18.txt: /api/issueListSubscription?fields=ticket).
-- No existing table is modified.

CREATE TABLE issue_list_subscriptions (
    id SERIAL PRIMARY KEY,
    ticket VARCHAR(255) UNIQUE NOT NULL,
    user_id VARCHAR(20) REFERENCES users(id) ON DELETE CASCADE,
    query TEXT NOT NULL DEFAULT '',
    subscribe BOOLEAN NOT NULL DEFAULT TRUE,
    context_type VARCHAR(100) NOT NULL DEFAULT 'Project',
    context_id VARCHAR(50) NOT NULL DEFAULT '0-0',
    folder_id VARCHAR(50),
    type VARCHAR(100) NOT NULL DEFAULT 'IssueListSubscriptionBean',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE issue_list_subscription_issues (
    id SERIAL PRIMARY KEY,
    subscription_id INT NOT NULL REFERENCES issue_list_subscriptions(id) ON DELETE CASCADE,
    issue_id VARCHAR(50) NOT NULL,
    matches BOOLEAN NOT NULL DEFAULT TRUE,
    ordinal INT NOT NULL DEFAULT 0
);

CREATE INDEX idx_issue_list_sub_issues ON issue_list_subscription_issues(subscription_id);

-- Seed a default ticket matching the request18.txt sample response.
INSERT INTO issue_list_subscriptions (ticket, query, subscribe, context_type, context_id)
VALUES ('g9qk4fe77jefpavefi1qacdhbefa3l', 'issue id: DEMO-4', TRUE, 'Project', '0-0');


-- #############################################################################
--  [10] الأصول المسموح بها
--  المصدر: migrations/000010_global_settings_allowed_origins.up.sql
-- #############################################################################

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



-- #############################################################################
--  [14] حقول مرشّح الأمان
--  المصدر: migrations/000014_security_filter_fields.up.sql
-- #############################################################################

CREATE TABLE security_filter_fields (
    id VARCHAR(100) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    entity_type VARCHAR(100) NOT NULL,
    field_type VARCHAR(100) DEFAULT 'SecurityFilterField'
);

-- إدراج البيانات للطلب رقم 30 (ProjectPeopleResponse)
INSERT INTO security_filter_fields (id, name, entity_type, field_type) VALUES
('role.scope', 'Scope', 'ProjectPeopleResponse', 'SecurityFilterField'),
('role.role', 'Role', 'ProjectPeopleResponse', 'SecurityFilterField'),
('role.permission', 'Permission', 'ProjectPeopleResponse', 'SecurityFilterField');

-- يمكن إضافة كيانات أخرى كما ظهرت في hh.json
INSERT INTO security_filter_fields (id, name, entity_type, field_type) VALUES
('user.login', 'Login', 'User', 'SecurityFilterField'),
('user.name', 'Name', 'User', 'SecurityFilterField'),
('group.name', 'Name', 'UserGroup', 'SecurityFilterField'),
('role.name', 'Name', 'Role', 'SecurityFilterField');


