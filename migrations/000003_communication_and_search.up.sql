-- =============================================================================
--  000003_communication_and_search.up.sql
--  =============================================================================
--  Purpose: Inbox threads, banners, and search assist subsystem tables.
--  Source: init_database.sql sections [5], [6], [7] (lines 1086-1310).
--  Dependencies: 000001_core_schema.
-- =============================================================================

-- #############################################################################
--  [5]  صندوق الوارد (4 جداول)
--  المصدر: migrations/000005_inbox_threads.up.sql
-- #############################################################################

-- Migration to add Inbox Threads and related entities

CREATE TABLE inbox_threads (
    id VARCHAR(50) PRIMARY KEY,
    user_id VARCHAR(20) REFERENCES users(id) ON DELETE CASCADE,
    read BOOLEAN DEFAULT FALSE,
    muted BOOLEAN DEFAULT FALSE,
    notified BOOLEAN DEFAULT TRUE,
    target_type VARCHAR(100),
    thread_id VARCHAR(100),
    timestamp BIGINT,
    updated BIGINT,
    subject_text TEXT,
    subject_target_id VARCHAR(100)
);

CREATE TABLE inbox_messages (
    id VARCHAR(50) PRIMARY KEY,
    thread_id VARCHAR(50) REFERENCES inbox_threads(id) ON DELETE CASCADE,
    author_id VARCHAR(20) REFERENCES users(id) ON DELETE SET NULL,
    author_group_id VARCHAR(20) REFERENCES user_groups(id) ON DELETE SET NULL,
    timestamp BIGINT,
    text TEXT,
    type VARCHAR(50),
    pseudo BOOLEAN DEFAULT FALSE,
    empty_field_text TEXT,
    target_id VARCHAR(100),
    target_type VARCHAR(100)
);

CREATE TABLE inbox_message_params (
    id SERIAL PRIMARY KEY,
    message_id VARCHAR(50) REFERENCES inbox_messages(id) ON DELETE CASCADE,
    key VARCHAR(255),
    value TEXT
);

CREATE TABLE inbox_activities (
    id VARCHAR(50) PRIMARY KEY,
    message_id VARCHAR(50) REFERENCES inbox_messages(id) ON DELETE CASCADE,
    category_id VARCHAR(100),
    added_id VARCHAR(100), -- Reference to User, Issue, Comment etc depending on type
    removed_id VARCHAR(100),
    target_id VARCHAR(100),
    target_type VARCHAR(100),
    timestamp BIGINT
);

-- Indexing for performance
CREATE INDEX idx_inbox_threads_user ON inbox_threads(user_id);
CREATE INDEX idx_inbox_messages_thread ON inbox_messages(thread_id);
CREATE INDEX idx_inbox_activities_message ON inbox_activities(message_id);


-- #############################################################################
--  [6]  اللافتات
--  المصدر: migrations/000006_banners.up.sql
-- #############################################################################

-- 000006_banners.up.sql
-- New independent schema for the banners block of GET /api/config
-- (request11.txt: fields=banners(globalBanner,globalBannerEnabled,systemEventsBanners)).
-- No existing table is modified.

CREATE TABLE banners_config (
    id SERIAL PRIMARY KEY,
    global_banner TEXT NOT NULL DEFAULT '',
    global_banner_enabled BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE system_events_banners (
    id SERIAL PRIMARY KEY,
    config_id INT REFERENCES banners_config(id) ON DELETE CASCADE,
    text TEXT NOT NULL DEFAULT ''
);

CREATE INDEX idx_system_events_banners_config ON system_events_banners(config_id);


-- #############################################################################
--  [7]  مساعدات البحث (6 جداول)
--  المصدر: migrations/000007_search_assist.up.sql
-- #############################################################################

CREATE TABLE search_assist_responses (
    id SERIAL PRIMARY KEY,
    query TEXT,
    caret INT,
    ignore_unresolved_setting BOOLEAN DEFAULT FALSE,
    type VARCHAR(100) DEFAULT 'SearchAssistResponse'
);

CREATE TABLE search_style_ranges (
    id SERIAL PRIMARY KEY,
    response_id INT REFERENCES search_assist_responses(id) ON DELETE CASCADE,
    length INT,
    start_pos INT,
    style VARCHAR(100),
    title TEXT,
    type VARCHAR(100) DEFAULT 'SearchStyleRange'
);

CREATE TABLE search_suggestions (
    id SERIAL PRIMARY KEY,
    response_id INT REFERENCES search_assist_responses(id) ON DELETE CASCADE,
    description TEXT,
    suggestion_group VARCHAR(255),
    icon VARCHAR(255),
    suggestion_option TEXT,
    prefix TEXT,
    suffix TEXT,
    class_name VARCHAR(255),
    matching_start INT,
    matching_end INT,
    caret INT,
    completion_start INT,
    completion_end INT,
    type VARCHAR(100) DEFAULT 'SearchSuggestion'
);

CREATE TABLE search_sort_fields (
    id VARCHAR(100) PRIMARY KEY,
    name VARCHAR(255),
    sortable_presentation VARCHAR(255),
    type VARCHAR(100) DEFAULT 'SortField'
);

CREATE TABLE search_sort_properties (
    id SERIAL PRIMARY KEY,
    response_id INT REFERENCES search_assist_responses(id) ON DELETE CASCADE,
    property_id VARCHAR(100),
    asc_order BOOLEAN DEFAULT TRUE,
    sort_field_id VARCHAR(100) REFERENCES search_sort_fields(id),
    type VARCHAR(100) DEFAULT 'SortProperty'
);

CREATE TABLE search_query_features (
    id SERIAL PRIMARY KEY,
    response_id INT REFERENCES search_assist_responses(id) ON DELETE CASCADE,
    contains_wildcard BOOLEAN DEFAULT FALSE,
    contains_latin BOOLEAN DEFAULT TRUE,
    contains_digits BOOLEAN DEFAULT FALSE,
    contains_not_latin BOOLEAN DEFAULT FALSE,
    number_of_words INT DEFAULT 0,
    has_and BOOLEAN DEFAULT FALSE,
    has_or BOOLEAN DEFAULT FALSE,
    has_sorting BOOLEAN DEFAULT FALSE,
    query_length INT DEFAULT 0,
    fields_count INT DEFAULT 0,
    user_in_fields BOOLEAN DEFAULT FALSE,
    me_in_fields BOOLEAN DEFAULT FALSE,
    date_period_in_fields BOOLEAN DEFAULT FALSE,
    sorted_by_relevance BOOLEAN DEFAULT FALSE,
    sort_by_relevance_setting BOOLEAN DEFAULT FALSE,
    experiment_version VARCHAR(50),
    experiment_group VARCHAR(50),
    user_left_experiment BOOLEAN DEFAULT FALSE,
    is_guest BOOLEAN DEFAULT FALSE,
    is_internal BOOLEAN DEFAULT TRUE,
    is_jb_team BOOLEAN DEFAULT FALSE,
    folder_selected BOOLEAN DEFAULT FALSE,
    project_folder_selected BOOLEAN DEFAULT FALSE,
    saved_query_folder_selected BOOLEAN DEFAULT FALSE,
    tag_folder_selected BOOLEAN DEFAULT FALSE,
    is_single_issue BOOLEAN DEFAULT FALSE,
    is_tree_view BOOLEAN DEFAULT TRUE,
    query_computation_time_ms BIGINT DEFAULT 0,
    time_since_last_search_ms BIGINT DEFAULT 0,
    recent_searches_count INT DEFAULT 0,
    last_search_query_text_similarity DOUBLE PRECISION DEFAULT 0,
    last_search_query_token_similarity DOUBLE PRECISION DEFAULT 0,
    predefined_field_comment_text_field BOOLEAN DEFAULT FALSE,
    predefined_field_work_text_field BOOLEAN DEFAULT FALSE,
    predefined_field_vcs_changes_field BOOLEAN DEFAULT FALSE,
    predefined_field_ticket_cc_groups_field BOOLEAN DEFAULT FALSE,
    predefined_field_mentions_field BOOLEAN DEFAULT FALSE,
    predefined_field_underestimation_field BOOLEAN DEFAULT FALSE,
    predefined_field_saved_query_field BOOLEAN DEFAULT FALSE,
    predefined_field_content_field BOOLEAN DEFAULT FALSE,
    predefined_field_project_field BOOLEAN DEFAULT FALSE,
    predefined_field_star_field BOOLEAN DEFAULT FALSE,
    predefined_field_attachment_name_field BOOLEAN DEFAULT FALSE,
    predefined_field_voted_by_field BOOLEAN DEFAULT FALSE,
    predefined_field_article_field BOOLEAN DEFAULT FALSE,
    predefined_field_reaction_from_field BOOLEAN DEFAULT FALSE,
    predefined_field_by_field BOOLEAN DEFAULT FALSE,
    predefined_field_commented_by_field BOOLEAN DEFAULT FALSE,
    predefined_field_custom_field BOOLEAN DEFAULT FALSE,
    predefined_field_code_field BOOLEAN DEFAULT FALSE,
    predefined_field_ticket_cc_field BOOLEAN DEFAULT FALSE,
    predefined_field_summary_field BOOLEAN DEFAULT FALSE,
    predefined_field_mentioned_in_field BOOLEAN DEFAULT FALSE,
    predefined_field_organization_field BOOLEAN DEFAULT FALSE,
    predefined_field_issue_field BOOLEAN DEFAULT FALSE,
    predefined_field_votes_field BOOLEAN DEFAULT FALSE,
    predefined_field_attachments_field BOOLEAN DEFAULT FALSE,
    predefined_field_links_field BOOLEAN DEFAULT FALSE,
    predefined_field_title_field BOOLEAN DEFAULT FALSE,
    predefined_field_work_field BOOLEAN DEFAULT FALSE,
    predefined_field_attachment_text_field BOOLEAN DEFAULT FALSE,
    predefined_field_comments_field BOOLEAN DEFAULT FALSE,
    predefined_field_document_type_field BOOLEAN DEFAULT FALSE,
    predefined_field_action_field BOOLEAN DEFAULT FALSE,
    predefined_field_has_field BOOLEAN DEFAULT FALSE,
    predefined_field_similar_to_field BOOLEAN DEFAULT FALSE,
    predefined_field_tag_field BOOLEAN DEFAULT FALSE,
    predefined_field_description_field BOOLEAN DEFAULT FALSE,
    predefined_field_visible_to_field BOOLEAN DEFAULT FALSE,
    predefined_field_sort_by_field BOOLEAN DEFAULT FALSE,
    predefined_field_commented_field BOOLEAN DEFAULT FALSE,
    predefined_field_article_author_field BOOLEAN DEFAULT FALSE,
    predefined_field_updated_by_field BOOLEAN DEFAULT FALSE,
    predefined_field_submitted_by_field BOOLEAN DEFAULT FALSE,
    predefined_field_resolved_field BOOLEAN DEFAULT FALSE,
    predefined_field_created_field BOOLEAN DEFAULT FALSE,
    predefined_field_updated_field BOOLEAN DEFAULT FALSE,
    type VARCHAR(100) DEFAULT 'SearchQueryFeatures'
);


