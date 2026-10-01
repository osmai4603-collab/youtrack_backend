-- =============================================================================
--  000002_user_profiles_and_widgets.up.sql
--  =============================================================================
--  Purpose: Extended user profile tables and dashboard widget extensions.
--  Source: init_database.sql sections [3], [4], and [15] (lines 957-1085, 1524-1535).
--  Dependencies: 000001_core_schema.
-- =============================================================================

-- #############################################################################
--  [3]  تفاصيل الملف الشخصي (6 جداول)
--  المصدر: migrations/000003_user_profiles_extended.up.sql
-- #############################################################################

-- 000003_user_profiles_extended.up.sql
-- Extended tables for storing individual user profile sub-components
--
-- ترقيعة: 000003 ينشئ ثمانية جداول، ستّة منها منقولة هنا. المستبعَدان هما
-- user_profile_tips (65 عمودًا لحالات "-shown" لتلميحات الواجهة) و
-- user_profile_ai (إعدادات محرّك المحادثة). لا يشير إليهما أي استعلام في
-- internal/؛ التطبيق يقرأ تفضيلات العرض من user_profile_appearance وحده.
CREATE TABLE IF NOT EXISTS user_profile_appearance (
    user_id VARCHAR(20) PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    open_cw_on_typing BOOLEAN DEFAULT TRUE,
    siv_sidebar_width INT DEFAULT 240,
    show_toolbar BOOLEAN DEFAULT TRUE,
    show_quick_view BOOLEAN DEFAULT TRUE,
    show_similar_issues BOOLEAN DEFAULT TRUE,
    show_knowledge_base_sidebar BOOLEAN DEFAULT TRUE,
    show_siv_sidebar BOOLEAN DEFAULT FALSE,
    dashboard_header_collapsed BOOLEAN DEFAULT FALSE,
    knowledge_base_subarticles_collapsed BOOLEAN DEFAULT FALSE,
    ai_link_suggestions_collapsed BOOLEAN DEFAULT FALSE,
    show_comments_in_activity_stream BOOLEAN DEFAULT TRUE,
    knowledge_base_sidebar_width INT DEFAULT 240,
    last_used_color VARCHAR(50) DEFAULT 'default',
    attachments_list_layout BOOLEAN DEFAULT FALSE,
    show_tooltips BOOLEAN DEFAULT TRUE,
    expand_navigation BOOLEAN DEFAULT TRUE,
    hide_comment_attachments BOOLEAN DEFAULT TRUE,
    sidebar_quick_view_mode BOOLEAN DEFAULT TRUE,
    expand_changes_in_activity_stream BOOLEAN DEFAULT FALSE,
    natural_comments_order BOOLEAN DEFAULT TRUE,
    use_absolute_dates BOOLEAN DEFAULT FALSE,
    use_markdown_editor BOOLEAN DEFAULT FALSE,
    show_inline_editor_toolbar BOOLEAN DEFAULT TRUE,
    show_vcs_changes_in_activity_stream BOOLEAN DEFAULT FALSE,
    issue_list_sidebar_width INT DEFAULT 240,
    attachments_collapsed BOOLEAN DEFAULT FALSE,
    use_summary_in_issue_links BOOLEAN DEFAULT TRUE,
    show_recent_entities BOOLEAN DEFAULT TRUE,
    exceptions_expanded BOOLEAN DEFAULT TRUE,
    onboarding_tour_panel_width INT DEFAULT 400,
    show_links_under_description BOOLEAN DEFAULT TRUE,
    recognized_text_sidebar_expanded BOOLEAN DEFAULT FALSE,
    quick_view_sidebar_width INT DEFAULT 240,
    modal_sidebar_width INT DEFAULT 240,
    show_sidebar_resizer_tip BOOLEAN DEFAULT TRUE,
    issues_table_view_mode BOOLEAN DEFAULT TRUE,
    attachments_sorting VARCHAR(50) DEFAULT 'default',
    hide_embedded_attachments BOOLEAN DEFAULT TRUE,
    compact_mode BOOLEAN DEFAULT FALSE,
    quick_view_width INT DEFAULT 0,
    show_history_in_activity_stream BOOLEAN DEFAULT FALSE,
    show_work_items_in_activity_stream BOOLEAN DEFAULT FALSE,
    first_day_of_week INT DEFAULT 0
);

CREATE TABLE IF NOT EXISTS user_profile_notifications (
    user_id VARCHAR(20) PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    show_unread_only BOOLEAN DEFAULT FALSE,
    mention_notifications_enabled BOOLEAN DEFAULT TRUE,
    duplicate_cluster_notifications_enabled BOOLEAN DEFAULT FALSE,
    show_system BOOLEAN DEFAULT FALSE,
    notify_on_own_changes BOOLEAN DEFAULT FALSE,
    auto_watch_on_field_set BOOLEAN DEFAULT TRUE,
    auto_watch_on_create BOOLEAN DEFAULT TRUE,
    auto_watch_on_comment BOOLEAN DEFAULT TRUE,
    auto_watch_on_update BOOLEAN DEFAULT FALSE,
    auto_watch_on_vote BOOLEAN DEFAULT TRUE,
    email_blocked BOOLEAN DEFAULT FALSE,
    email_block_reason TEXT,
    email_notifications_enabled BOOLEAN DEFAULT TRUE,
    use_plain_text_emails BOOLEAN DEFAULT FALSE,
    disabled_direct BOOLEAN DEFAULT FALSE,
    disabled_subscription BOOLEAN DEFAULT FALSE,
    disabled_system BOOLEAN DEFAULT FALSE,
    mailbox_integration_notifications_enabled BOOLEAN DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS user_profile_helpdesk (
    user_id VARCHAR(20) PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    is_reporter BOOLEAN DEFAULT FALSE,
    is_agent BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS user_profile_helpdesk_projects (
    user_id VARCHAR(20) REFERENCES users(id) ON DELETE CASCADE,
    project_id VARCHAR(20) REFERENCES projects(id) ON DELETE CASCADE,
    role_type VARCHAR(20) NOT NULL, -- 'agent' or 'reporter'
    PRIMARY KEY (user_id, project_id, role_type)
);

CREATE TABLE IF NOT EXISTS user_profile_grazie (
    user_id VARCHAR(20) PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    has_more_tokens BOOLEAN DEFAULT TRUE,
    enabled BOOLEAN DEFAULT FALSE,
    excluded_issue_types TEXT DEFAULT '',
    cycle_restart BIGINT DEFAULT 1788271206218,
    enable_spell_checker BOOLEAN DEFAULT TRUE,
    spell_checker_enabled_in_system BOOLEAN DEFAULT TRUE,
    free_license BOOLEAN DEFAULT FALSE,
    enable_text_completion BOOLEAN DEFAULT TRUE,
    text_completion_enabled_in_system BOOLEAN DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS user_profile_questionnaire (
    user_id VARCHAR(20) PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    show_survey BOOLEAN DEFAULT FALSE,
    show_pmf_survey BOOLEAN DEFAULT FALSE,
    demo_eligibility_timestamp BIGINT
);


-- #############################################################################
--  [4]  أحجام widgets المتوقعة
--  المصدر: migrations/000004_dashboard_widgets_extended.up.sql
-- #############################################################################

-- 000004_dashboard_widgets_extended.up.sql
-- Add the expectedHeight/expectedWidth columns used by the
-- GET /api/admin/widgets/general endpoint (request6.txt).

ALTER TABLE dashboard_widgets
    ADD COLUMN IF NOT EXISTS expected_height VARCHAR(20),
    ADD COLUMN IF NOT EXISTS expected_width VARCHAR(20);



-- #############################################################################
--  [15] نمط التاريخ
--  المصدر: migrations/000015_user_profile_date_field_pattern.up.sql
-- #############################################################################

-- 000015_user_profile_date_field_pattern.up.sql
-- Add date_field_pattern column to store the datetime format (e.g., "d MMM yyyy HH:mm")
-- The existing date_pattern column stores the date-only format (e.g., "d MMM yyyy")

ALTER TABLE user_profiles ADD COLUMN date_field_pattern VARCHAR(50);


