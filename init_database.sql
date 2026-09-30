-- Consolidated YouTrack schema.
-- Source: /home/osm/StudioProjects/youtrack_backend/db/migrations/*.up.sql
-- Ordered from foundational tables to dependent feature tables.

-- -----------------------------------------------------------------------------
-- Source migration: 000001_init_schema.up.sql
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS field_styles (
    id VARCHAR(20) PRIMARY KEY,
    background VARCHAR(20),
    foreground VARCHAR(20)
);

CREATE TABLE IF NOT EXISTS project_types (
    id VARCHAR(50) PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS user_types (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL
);

CREATE TABLE IF NOT EXISTS organizations (
    id VARCHAR(20) PRIMARY KEY,
    key VARCHAR(100),
    name VARCHAR(255),
    icon_url TEXT,
    projects_count INT DEFAULT 0,
    description TEXT,
    audit_target_id VARCHAR(100)
);

CREATE TABLE IF NOT EXISTS permissions (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255),
    description TEXT,
    permission_entity_type VARCHAR(100),
    localized_permission_entity_type VARCHAR(100),
    operation VARCHAR(50),
    is_global BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS work_time_settings (
    id SERIAL PRIMARY KEY,
    minutes_a_day INT DEFAULT 480,
    minutes_a_day_presentation VARCHAR(10) DEFAULT '8h',
    days_a_week INT DEFAULT 5,
    first_day_of_week INT DEFAULT 0
);

CREATE TABLE IF NOT EXISTS work_days (
    settings_id INT REFERENCES work_time_settings(id) ON DELETE CASCADE,
    day_number INT,
    PRIMARY KEY (settings_id, day_number)
);

CREATE TABLE IF NOT EXISTS work_item_types (
    id VARCHAR(20) PRIMARY KEY,
    name VARCHAR(255),
    color_id VARCHAR(20) REFERENCES field_styles(id) ON DELETE SET NULL,
    auto_attach BOOLEAN DEFAULT FALSE,
    description TEXT
);

CREATE TABLE IF NOT EXISTS services (
    id VARCHAR(100) PRIMARY KEY,
    name VARCHAR(255),
    key VARCHAR(255),
    home_url TEXT,
    application_name VARCHAR(255),
    vendor VARCHAR(255),
    version VARCHAR(50),
    trusted BOOLEAN DEFAULT FALSE,
    icon_url TEXT,
    user_uri_pattern TEXT,
    group_uri_pattern TEXT,
    audience VARCHAR(255),
    immutable BOOLEAN DEFAULT FALSE,
    client_credentials_flow_enabled BOOLEAN DEFAULT FALSE,
    auth_code_flow_enabled BOOLEAN DEFAULT FALSE,
    implicit_flow_enabled BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS vcs_hosting_servers (
    id VARCHAR(20) PRIMARY KEY,
    url TEXT,
    server_type VARCHAR(50),
    is_predefined BOOLEAN DEFAULT TRUE,
    app_id VARCHAR(100),
    app_name VARCHAR(255),
    application_id VARCHAR(100),
    ssl_key_id VARCHAR(20)
);

CREATE TABLE IF NOT EXISTS feature_flags (
    id VARCHAR(255) PRIMARY KEY,
    enabled BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS field_types (
    id VARCHAR(50) PRIMARY KEY,
    presentation VARCHAR(100),
    value_type VARCHAR(50),
    is_bundle_type BOOLEAN DEFAULT FALSE,
    is_multi_value BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS field_converters (
    id VARCHAR(100) PRIMARY KEY,
    from_type_id VARCHAR(50) REFERENCES field_types(id) ON DELETE CASCADE,
    to_type_id VARCHAR(50) REFERENCES field_types(id) ON DELETE CASCADE,
    localized_name VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS bundles (
    id VARCHAR(20) PRIMARY KEY,
    bundle_type VARCHAR(50) NOT NULL,
    name VARCHAR(255),
    is_updateable BOOLEAN DEFAULT FALSE
);

-- 2. Core Entities with FKs
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(20) PRIMARY KEY,
    login VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(255) UNIQUE,
    full_name VARCHAR(255),
    name VARCHAR(255),
    avatar_url TEXT,
    user_type_id VARCHAR(50) REFERENCES user_types(id) ON DELETE SET NULL,
    is_email_verified BOOLEAN DEFAULT FALSE,
    guest BOOLEAN DEFAULT FALSE,
    online BOOLEAN DEFAULT FALSE,
    banned BOOLEAN DEFAULT FALSE,
    ban_badge TEXT,
    password_hash TEXT,
    can_read_profile BOOLEAN DEFAULT TRUE,
    is_locked BOOLEAN DEFAULT FALSE,
    ring_id VARCHAR(100)
);

CREATE TABLE IF NOT EXISTS user_profiles (
    user_id VARCHAR(20) PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    timezone_id VARCHAR(100),
    locale_id VARCHAR(20),
    date_pattern VARCHAR(50),
    date_field_pattern VARCHAR(50),
    email_notifications_enabled BOOLEAN DEFAULT TRUE,
    mention_notifications_enabled BOOLEAN DEFAULT TRUE,
    auto_watch_on_create BOOLEAN DEFAULT TRUE,
    auto_watch_on_comment BOOLEAN DEFAULT TRUE,
    compact_mode BOOLEAN DEFAULT FALSE,
    expand_navigation BOOLEAN DEFAULT TRUE,
    natural_comments_order BOOLEAN DEFAULT TRUE,
    use_markdown_editor BOOLEAN DEFAULT FALSE,
    show_sidebar BOOLEAN DEFAULT TRUE,
    unresolved_issues_only BOOLEAN DEFAULT FALSE,
    period_format_id VARCHAR(20),
    is_time_tracking_available BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS project_teams (
    id VARCHAR(20) PRIMARY KEY,
    name VARCHAR(255),
    project_id VARCHAR(20),
    description TEXT,
    icon TEXT,
    audit_target_id VARCHAR(100),
    all_users_group BOOLEAN DEFAULT FALSE,
    is_updatable BOOLEAN DEFAULT TRUE,
    is_removable BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS projects (
    id VARCHAR(20) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    short_name VARCHAR(50) NOT NULL UNIQUE,
    project_type_id VARCHAR(50) REFERENCES project_types(id) ON DELETE SET NULL,
    pinned BOOLEAN DEFAULT FALSE,
    icon_url TEXT,
    template BOOLEAN DEFAULT FALSE,
    archived BOOLEAN DEFAULT FALSE,
    restricted BOOLEAN DEFAULT FALSE,
    has_articles BOOLEAN DEFAULT FALSE,
    is_demo BOOLEAN DEFAULT FALSE,
    fields_sorted BOOLEAN DEFAULT FALSE,
    query TEXT,
    issues_url TEXT,
    description TEXT,
    leader_id VARCHAR(20) REFERENCES users(id) ON DELETE SET NULL,
    team_id VARCHAR(20) REFERENCES project_teams(id) ON DELETE SET NULL,
    organization_id VARCHAR(20) REFERENCES organizations(id) ON DELETE SET NULL,
    creation_time BIGINT,
    from_email VARCHAR(255),
    from_personal VARCHAR(255),
    reply_to_email VARCHAR(255),
    email_delimiter TEXT,
    use_email_delimiter BOOLEAN DEFAULT FALSE,
    supports_email_delimiter BOOLEAN DEFAULT FALSE,
    default_visibility_group_id VARCHAR(20),
    default_smtp BOOLEAN DEFAULT FALSE,
    source_template VARCHAR(20),
    audit_target_id VARCHAR(100),
    historical_short_names TEXT[],
    created_at BIGINT,
    updated_at BIGINT
);

CREATE TABLE IF NOT EXISTS user_groups (
    id VARCHAR(20) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    group_type VARCHAR(50),
    all_users_group BOOLEAN DEFAULT FALSE,
    icon TEXT,
    description TEXT,
    audit_target_id VARCHAR(100),
    is_updatable BOOLEAN DEFAULT FALSE,
    is_removable BOOLEAN DEFAULT FALSE,
    team_for_project_id VARCHAR(20) REFERENCES projects(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS user_group_members (
    user_id VARCHAR(20) REFERENCES users(id) ON DELETE CASCADE,
    group_id VARCHAR(20) REFERENCES user_groups(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, group_id)
);

CREATE TABLE IF NOT EXISTS project_team_members (
    team_id VARCHAR(20) REFERENCES project_teams(id) ON DELETE CASCADE,
    user_id VARCHAR(20) REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (team_id, user_id)
);

-- 3. Custom Fields & Bundles
CREATE TABLE IF NOT EXISTS custom_fields (
    id VARCHAR(20) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    localized_name VARCHAR(255),
    ordinal INT DEFAULT 0,
    aliases TEXT,
    field_type_id VARCHAR(50) REFERENCES field_types(id) ON DELETE SET NULL,
    has_running_job BOOLEAN DEFAULT FALSE,
    is_auto_attached BOOLEAN DEFAULT FALSE,
    is_displayed_in_issue_list BOOLEAN DEFAULT FALSE,
    is_updateable BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS project_custom_fields (
    id VARCHAR(20) PRIMARY KEY,
    project_id VARCHAR(20) REFERENCES projects(id) ON DELETE CASCADE,
    custom_field_id VARCHAR(20) REFERENCES custom_fields(id) ON DELETE CASCADE,
    bundle_id VARCHAR(20) REFERENCES bundles(id) ON DELETE SET NULL,
    field_type VARCHAR(80),
    can_be_empty BOOLEAN DEFAULT TRUE,
    empty_field_text VARCHAR(255),
    has_running_job BOOLEAN DEFAULT FALSE,
    ordinal INT DEFAULT 0,
    is_spent_time BOOLEAN DEFAULT FALSE,
    is_estimation BOOLEAN DEFAULT FALSE,
    is_public BOOLEAN DEFAULT TRUE,
    is_sla_field BOOLEAN DEFAULT FALSE,
    size INT,
    empty_value_equals_element BOOLEAN,
    has_state_machine BOOLEAN DEFAULT FALSE,
    visible_to_group_id VARCHAR(20) REFERENCES user_groups(id) ON DELETE SET NULL,
    updatable_by_group_id VARCHAR(20) REFERENCES user_groups(id) ON DELETE SET NULL,
    values_condition_field_id VARCHAR(20)
);

CREATE TABLE IF NOT EXISTS bundle_values (
    id VARCHAR(20) PRIMARY KEY,
    bundle_id VARCHAR(20) REFERENCES bundles(id) ON DELETE CASCADE,
    name VARCHAR(255),
    description TEXT,
    localized_name VARCHAR(255),
    is_resolved BOOLEAN,
    archived BOOLEAN DEFAULT FALSE,
    color_id VARCHAR(20) REFERENCES field_styles(id) ON DELETE SET NULL,
    ordinal INT DEFAULT 0,
    release_date BIGINT,
    released BOOLEAN,
    build_integration TEXT,
    build_link TEXT,
    assemble_date BIGINT,
    owner_id VARCHAR(20) REFERENCES users(id) ON DELETE SET NULL,
    login VARCHAR(100),
    avatar_url TEXT,
    auto_attach BOOLEAN DEFAULT FALSE,
    show_localized_name_in_admin BOOLEAN DEFAULT FALSE,
    has_running_job BOOLEAN DEFAULT FALSE,
    start_date BIGINT,
    users_count INT,
    ring_id VARCHAR(100)
);

-- 4. Issues & Articles
CREATE TABLE IF NOT EXISTS saved_queries (
    id VARCHAR(20) PRIMARY KEY,
    name VARCHAR(255),
    query TEXT,
    folder_id VARCHAR(255),
    is_updatable BOOLEAN DEFAULT FALSE,
    is_deletable BOOLEAN DEFAULT FALSE,
    is_shareable BOOLEAN DEFAULT FALSE,
    pinned BOOLEAN DEFAULT FALSE,
    pinned_by_default BOOLEAN DEFAULT FALSE,
    pinned_in_helpdesk BOOLEAN DEFAULT FALSE,
    issues_url TEXT,
    owner_id VARCHAR(20) REFERENCES users(id) ON DELETE SET NULL,
    sort_order_sortable BOOLEAN DEFAULT FALSE,
    visible_for_id VARCHAR(20) REFERENCES user_groups(id) ON DELETE SET NULL,
    updateable_by_id VARCHAR(20) REFERENCES user_groups(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS issues (
    id VARCHAR(20) PRIMARY KEY,
    id_readable VARCHAR(50) NOT NULL,
    number_in_project INT,
    summary TEXT NOT NULL,
    description TEXT,
    project_id VARCHAR(20) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    reporter_id VARCHAR(20) REFERENCES users(id) ON DELETE SET NULL,
    creator_id VARCHAR(20) REFERENCES users(id) ON DELETE SET NULL,
    updater_id VARCHAR(20) REFERENCES users(id) ON DELETE SET NULL,
    created BIGINT,
    updated BIGINT,
    resolved BIGINT,
    votes INT DEFAULT 0,
    is_draft BOOLEAN DEFAULT FALSE,
    unauthenticated_reporter BOOLEAN DEFAULT FALSE,
    can_undo_comment BOOLEAN DEFAULT FALSE,
    can_add_public_comment BOOLEAN DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS comments (
    id VARCHAR(20) PRIMARY KEY,
    issue_id VARCHAR(20) NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    author_id VARCHAR(20) REFERENCES users(id) ON DELETE SET NULL,
    text TEXT,
    created BIGINT,
    updated BIGINT,
    url TEXT,
    is_deleted BOOLEAN DEFAULT FALSE,
    visibility_type VARCHAR(50),
    parent_comment_id VARCHAR(20) REFERENCES comments(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS articles (
    id VARCHAR(20) PRIMARY KEY,
    id_readable VARCHAR(50),
    summary TEXT NOT NULL,
    project_id VARCHAR(20) REFERENCES projects(id) ON DELETE CASCADE,
    reporter_id VARCHAR(20) REFERENCES users(id) ON DELETE SET NULL,
    parent_article_id VARCHAR(20) REFERENCES articles(id) ON DELETE SET NULL,
    ordinal INT DEFAULT 0,
    has_unpublished_changes BOOLEAN DEFAULT FALSE,
    has_children BOOLEAN DEFAULT FALSE,
    has_star BOOLEAN DEFAULT FALSE,
    is_updatable BOOLEAN DEFAULT FALSE,
    is_deletable BOOLEAN DEFAULT FALSE,
    collaborative_draft_id VARCHAR(100),
    updated BIGINT
);

CREATE TABLE IF NOT EXISTS attachments (
    id VARCHAR(20) PRIMARY KEY,
    issue_id VARCHAR(20) REFERENCES issues(id) ON DELETE CASCADE,
    comment_id VARCHAR(20) REFERENCES comments(id) ON DELETE CASCADE,
    article_id VARCHAR(20) REFERENCES articles(id) ON DELETE CASCADE,
    author_id VARCHAR(20) REFERENCES users(id) ON DELETE SET NULL,
    name VARCHAR(500) NOT NULL,
    size BIGINT,
    mime_type VARCHAR(255),
    url TEXT,
    created BIGINT,
    visibility_type VARCHAR(50)
);

CREATE TABLE IF NOT EXISTS work_items (
    id VARCHAR(20) PRIMARY KEY,
    issue_id VARCHAR(20) NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    author_id VARCHAR(20) REFERENCES users(id) ON DELETE SET NULL,
    creator_id VARCHAR(20) REFERENCES users(id) ON DELETE SET NULL,
    work_type_id VARCHAR(20) REFERENCES work_item_types(id) ON DELETE SET NULL,
    duration_minutes INT,
    date BIGINT,
    text TEXT,
    created BIGINT,
    updated BIGINT
);

CREATE TABLE IF NOT EXISTS issue_links (
    id VARCHAR(20) PRIMARY KEY,
    source_issue_id VARCHAR(20) NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    target_issue_id VARCHAR(20) NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    link_type VARCHAR(100),
    direction VARCHAR(20)
);

CREATE TABLE IF NOT EXISTS tags (
    id VARCHAR(20) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    color_id VARCHAR(20) REFERENCES field_styles(id) ON DELETE SET NULL,
    is_deletable BOOLEAN DEFAULT FALSE,
    is_updatable BOOLEAN DEFAULT FALSE,
    is_usable BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS issue_tags (
    issue_id VARCHAR(20) REFERENCES issues(id) ON DELETE CASCADE,
    tag_id VARCHAR(20) REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (issue_id, tag_id)
);

CREATE TABLE IF NOT EXISTS issue_custom_field_values (
    id VARCHAR(20) PRIMARY KEY,
    issue_id VARCHAR(20) NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    project_custom_field_id VARCHAR(20) REFERENCES project_custom_fields(id) ON DELETE CASCADE,
    name VARCHAR(255),
    field_type VARCHAR(50),
    has_state_machine BOOLEAN DEFAULT FALSE,
    is_updatable BOOLEAN DEFAULT FALSE,
    visible_on_list BOOLEAN DEFAULT FALSE,
    paused_time BIGINT
);

CREATE TABLE IF NOT EXISTS issue_field_enum_values (
    id VARCHAR(20) PRIMARY KEY,
    name VARCHAR(255),
    localized_name VARCHAR(255),
    description TEXT,
    is_resolved BOOLEAN,
    color_id VARCHAR(20) REFERENCES field_styles(id) ON DELETE SET NULL,
    login VARCHAR(100),
    email VARCHAR(255),
    full_name VARCHAR(255),
    avatar_url TEXT,
    presentation VARCHAR(255),
    minutes INT,
    text TEXT,
    build_integration TEXT,
    build_link TEXT,
    archived BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS issue_voters (
    issue_id VARCHAR(20) REFERENCES issues(id) ON DELETE CASCADE,
    user_id VARCHAR(20) REFERENCES users(id) ON DELETE CASCADE,
    has_vote BOOLEAN DEFAULT FALSE,
    PRIMARY KEY (issue_id, user_id)
);

CREATE TABLE IF NOT EXISTS issue_watchers (
    issue_id VARCHAR(20) REFERENCES issues(id) ON DELETE CASCADE,
    user_id VARCHAR(20) REFERENCES users(id) ON DELETE CASCADE,
    has_star BOOLEAN DEFAULT FALSE,
    PRIMARY KEY (issue_id, user_id)
);

-- 5. Visibility
CREATE TABLE IF NOT EXISTS visibility (
    id VARCHAR(20) PRIMARY KEY,
    visibility_type VARCHAR(50),
    entity_type VARCHAR(50),
    entity_id VARCHAR(20)
);

CREATE TABLE IF NOT EXISTS visibility_permitted_users (
    visibility_id VARCHAR(20) REFERENCES visibility(id) ON DELETE CASCADE,
    user_id VARCHAR(20) REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (visibility_id, user_id)
);

CREATE TABLE IF NOT EXISTS visibility_permitted_groups (
    visibility_id VARCHAR(20) REFERENCES visibility(id) ON DELETE CASCADE,
    group_id VARCHAR(20) REFERENCES user_groups(id) ON DELETE CASCADE,
    PRIMARY KEY (visibility_id, group_id)
);

-- 6. Recents & Inbox
CREATE TABLE IF NOT EXISTS recent_issues (
    id VARCHAR(20) PRIMARY KEY,
    user_id VARCHAR(20) REFERENCES users(id) ON DELETE CASCADE,
    issue_id VARCHAR(20) REFERENCES issues(id) ON DELETE CASCADE,
    pinned BOOLEAN DEFAULT FALSE,
    date BIGINT
);

CREATE TABLE IF NOT EXISTS recent_articles (
    id VARCHAR(20) PRIMARY KEY,
    user_id VARCHAR(20) REFERENCES users(id) ON DELETE CASCADE,
    article_id VARCHAR(20) REFERENCES articles(id) ON DELETE CASCADE,
    pinned BOOLEAN DEFAULT FALSE,
    date BIGINT
);

CREATE TABLE IF NOT EXISTS inbox_folders (
    id VARCHAR(50) PRIMARY KEY,
    user_id VARCHAR(20) REFERENCES users(id) ON DELETE CASCADE,
    last_seen BIGINT DEFAULT 0,
    last_notified BIGINT DEFAULT 0,
    enabled BOOLEAN DEFAULT TRUE
);

-- 7. Roles & Permissions Cache
CREATE TABLE IF NOT EXISTS roles (
    id VARCHAR(100) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    audit_target_id VARCHAR(100),
    is_updatable BOOLEAN DEFAULT FALSE,
    immutable BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id VARCHAR(100) REFERENCES roles(id) ON DELETE CASCADE,
    permission_id VARCHAR(255) REFERENCES permissions(id) ON DELETE CASCADE,
    name VARCHAR(255),
    description TEXT,
    permission_entity_type VARCHAR(100),
    localized_permission_entity_type VARCHAR(100),
    operation VARCHAR(50),
    is_global BOOLEAN DEFAULT FALSE,
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE IF NOT EXISTS permission_dependencies (
    permission_id VARCHAR(255) REFERENCES permissions(id) ON DELETE CASCADE,
    dependent_permission_id VARCHAR(255) REFERENCES permissions(id) ON DELETE CASCADE,
    dependent_permission_name VARCHAR(255),
    PRIMARY KEY (permission_id, dependent_permission_id)
);

CREATE TABLE IF NOT EXISTS assigned_roles (
    id VARCHAR(20) PRIMARY KEY,
    role_id VARCHAR(100) REFERENCES roles(id) ON DELETE CASCADE,
    audit_target_id VARCHAR(100),
    holder_type VARCHAR(50),
    holder_id VARCHAR(20),
    holder_name VARCHAR(255),
    scope_type VARCHAR(50),
    scope_id VARCHAR(20),
    scope_project_id VARCHAR(20) REFERENCES projects(id) ON DELETE CASCADE,
    scope_organization_id VARCHAR(20) REFERENCES organizations(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS cached_permissions (
    id VARCHAR(100) REFERENCES permissions(id) ON DELETE CASCADE,
    user_id VARCHAR(20) REFERENCES users(id) ON DELETE CASCADE,
    is_global BOOLEAN DEFAULT FALSE,
    PRIMARY KEY (id, user_id)
);

CREATE TABLE IF NOT EXISTS cached_permission_projects (
    permission_id VARCHAR(100) REFERENCES permissions(id) ON DELETE CASCADE,
    project_id VARCHAR(20) REFERENCES projects(id) ON DELETE CASCADE,
    PRIMARY KEY (permission_id, project_id)
);

-- 8. Agile & Boards
CREATE TABLE IF NOT EXISTS agile_boards (
    id VARCHAR(20) PRIMARY KEY,
    name VARCHAR(255),
    is_demo BOOLEAN DEFAULT FALSE,
    is_updatable BOOLEAN DEFAULT FALSE,
    favorite BOOLEAN DEFAULT FALSE,
    flat_backlog BOOLEAN DEFAULT FALSE,
    hide_orphans_swimlane BOOLEAN DEFAULT FALSE,
    orphans_at_the_top BOOLEAN DEFAULT FALSE,
    colorize_custom_fields BOOLEAN DEFAULT FALSE,
    card_on_several_sprints BOOLEAN DEFAULT FALSE,
    owner_id VARCHAR(20) REFERENCES users(id) ON DELETE SET NULL,
    estimation_field_id VARCHAR(20) REFERENCES custom_fields(id) ON DELETE SET NULL,
    original_estimation_field_id VARCHAR(20) REFERENCES custom_fields(id) ON DELETE SET NULL,
    backlog_id VARCHAR(20) REFERENCES saved_queries(id) ON DELETE SET NULL,
    created_with_original BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS agile_board_projects (
    agile_id VARCHAR(20) REFERENCES agile_boards(id) ON DELETE CASCADE,
    project_id VARCHAR(20) REFERENCES projects(id) ON DELETE CASCADE,
    PRIMARY KEY (agile_id, project_id)
);

CREATE TABLE IF NOT EXISTS sprints (
    id VARCHAR(20) PRIMARY KEY,
    agile_id VARCHAR(20) REFERENCES agile_boards(id) ON DELETE CASCADE,
    name VARCHAR(255),
    start BIGINT,
    finish BIGINT,
    goal TEXT,
    ordinal INT DEFAULT 0,
    archived BOOLEAN DEFAULT FALSE,
    is_default BOOLEAN DEFAULT FALSE,
    is_started BOOLEAN DEFAULT FALSE,
    report_id VARCHAR(20)
);

CREATE TABLE IF NOT EXISTS board_columns (
    id VARCHAR(20) PRIMARY KEY,
    agile_id VARCHAR(20) REFERENCES agile_boards(id) ON DELETE CASCADE,
    collapsed BOOLEAN DEFAULT FALSE,
    ordinal INT DEFAULT 0,
    is_resolved BOOLEAN DEFAULT FALSE,
    is_visible BOOLEAN DEFAULT TRUE,
    color_id VARCHAR(20) REFERENCES field_styles(id) ON DELETE SET NULL,
    parent_id VARCHAR(20) REFERENCES board_columns(id) ON DELETE SET NULL,
    wip_limit_min INT,
    wip_limit_max INT
);

CREATE TABLE IF NOT EXISTS board_column_field_values (
    id VARCHAR(20) PRIMARY KEY,
    column_id VARCHAR(20) REFERENCES board_columns(id) ON DELETE CASCADE,
    name VARCHAR(255),
    presentation VARCHAR(255),
    ordinal INT DEFAULT 0,
    is_resolved BOOLEAN DEFAULT FALSE,
    can_update BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS board_cells (
    id VARCHAR(50) PRIMARY KEY,
    column_id VARCHAR(20) REFERENCES board_columns(id) ON DELETE CASCADE,
    row_id VARCHAR(50),
    issues_count INT DEFAULT 0,
    too_many_issues BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS swimlane_settings (
    id VARCHAR(20) PRIMARY KEY,
    agile_id VARCHAR(20) REFERENCES agile_boards(id) ON DELETE CASCADE,
    enabled BOOLEAN DEFAULT FALSE,
    field_id VARCHAR(20) REFERENCES custom_fields(id) ON DELETE SET NULL,
    default_card_type VARCHAR(50)
);

-- 9. Apps & Dashboards
CREATE TABLE IF NOT EXISTS apps (
    id VARCHAR(20) PRIMARY KEY,
    name VARCHAR(500) NOT NULL,
    title VARCHAR(255),
    version VARCHAR(50),
    language_id VARCHAR(10),
    icon TEXT,
    dark_icon TEXT,
    auto_attach BOOLEAN DEFAULT FALSE,
    permanent BOOLEAN DEFAULT FALSE,
    from_marketplace BOOLEAN DEFAULT FALSE,
    has_widget_or_http  BOOLEAN DEFAULT FALSE,
    can_be_attached BOOLEAN DEFAULT FALSE,
    can_update_content BOOLEAN DEFAULT TRUE,
    can_delete_content BOOLEAN DEFAULT TRUE,
    has_broken_usages BOOLEAN DEFAULT FALSE,
    model TEXT,
    updated BIGINT,
    updated_by_id VARCHAR(20) REFERENCES users(id) ON DELETE SET NULL,
    marketplace_id INT,
    vendor_name VARCHAR(255),
    vendor_url TEXT,
    vendor_email VARCHAR(255),
    global_config_enabled BOOLEAN DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS app_tags (
    app_id VARCHAR(20) REFERENCES apps(id) ON DELETE CASCADE,
    name VARCHAR(100),
    PRIMARY KEY (app_id, name)
);

CREATE TABLE IF NOT EXISTS project_app_configurations (
    id VARCHAR(20) PRIMARY KEY,
    project_id VARCHAR(20) REFERENCES projects(id) ON DELETE CASCADE,
    app_id VARCHAR(20) REFERENCES apps(id) ON DELETE CASCADE,
    enabled BOOLEAN DEFAULT TRUE,
    is_broken BOOLEAN DEFAULT FALSE,
    can_update BOOLEAN DEFAULT FALSE,
    missing_required_settings BOOLEAN DEFAULT FALSE,
    project_settings TEXT
);

CREATE TABLE IF NOT EXISTS dashboard_widgets (
    id VARCHAR(20) PRIMARY KEY,
    name VARCHAR(255),
    key VARCHAR(100),
    app_id VARCHAR(20),
    app_name VARCHAR(255),
    app_title VARCHAR(255),
    description TEXT,
    extension_point VARCHAR(50),
    icon_path TEXT,
    index_path TEXT,
    configurable BOOLEAN DEFAULT FALSE,
    collapsed BOOLEAN DEFAULT FALSE,
    borderless BOOLEAN DEFAULT FALSE,
    show_header BOOLEAN DEFAULT TRUE,
    default_height VARCHAR(20),
    default_width VARCHAR(20),
    vendor_name VARCHAR(255),
    vendor_email VARCHAR(255),
    vendor_url TEXT,
    marketplace_id INT,
    app_icon_path TEXT,
    app_dark_icon_path TEXT,
    guard TEXT,
    expected_height VARCHAR(20),
    expected_width VARCHAR(20)
);

CREATE TABLE IF NOT EXISTS project_dashboard_widgets (
    id VARCHAR(20) PRIMARY KEY,
    project_id VARCHAR(20) REFERENCES projects(id) ON DELETE CASCADE,
    widget_id VARCHAR(20) REFERENCES dashboard_widgets(id) ON DELETE CASCADE,
    key VARCHAR(10),
    x INT DEFAULT 0,
    y INT DEFAULT 0,
    width INT DEFAULT 1,
    height INT DEFAULT 1,
    settings TEXT
);

CREATE TABLE IF NOT EXISTS project_time_tracking_settings (
    id VARCHAR(20) PRIMARY KEY,
    project_id VARCHAR(20) REFERENCES projects(id) ON DELETE CASCADE,
    enabled BOOLEAN DEFAULT FALSE,
    estimate_field_id VARCHAR(20) REFERENCES project_custom_fields(id) ON DELETE SET NULL,
    time_spent_field_id VARCHAR(20) REFERENCES project_custom_fields(id) ON DELETE SET NULL
);

-- 10. Global & Notification Settings
CREATE TABLE IF NOT EXISTS global_settings (
    id SERIAL PRIMARY KEY,
    allow_all_origins BOOLEAN DEFAULT FALSE,
    image_text_recognition_enabled BOOLEAN DEFAULT TRUE,
    ocr_supported VARCHAR(10),
    email_settings_enabled BOOLEAN DEFAULT TRUE,
    email_settings_is_default BOOLEAN DEFAULT TRUE,
    version VARCHAR(20),
    build VARCHAR(20),
    release_date BIGINT,
    default_page VARCHAR(100),
    context_path VARCHAR(100),
    read_only BOOLEAN DEFAULT FALSE,
    statistics_enabled BOOLEAN DEFAULT TRUE,
    helpdesk_enabled BOOLEAN DEFAULT FALSE,
    max_upload_file_size BIGINT,
    max_export_items INT,
    sse_ping_timeout_ms BIGINT
);

CREATE TABLE IF NOT EXISTS filter_fields (
    id VARCHAR(100) PRIMARY KEY,
    name VARCHAR(255),
    field_type VARCHAR(50),
    custom_field_id VARCHAR(20) REFERENCES custom_fields(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS notification_templates (
    id VARCHAR(255) PRIMARY KEY,
    file_name VARCHAR(255),
    description TEXT,
    content TEXT NOT NULL,
    overrided BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS notification_template_groups (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255),
    title VARCHAR(255),
    description TEXT,
    enabled BOOLEAN DEFAULT TRUE,
    customizable BOOLEAN DEFAULT FALSE,
    parent_group_id VARCHAR(255) REFERENCES notification_template_groups(id) ON DELETE SET NULL
);

-- -----------------------------------------------------------------------------
-- Source migration: 000003_user_profiles_extended.up.sql
-- -----------------------------------------------------------------------------
-- 000003_user_profiles_extended.up.sql
-- Extended tables for storing individual user profile sub-components

CREATE TABLE IF NOT EXISTS user_profile_tips (
    user_id VARCHAR(20) PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    issue_list_page_completed BOOLEAN DEFAULT FALSE,
    issue_page_completed BOOLEAN DEFAULT FALSE,
    helpdesk_team_tip_shown BOOLEAN DEFAULT FALSE,
    project_overview_tip_shown BOOLEAN DEFAULT FALSE,
    project_settings_tip_shown BOOLEAN DEFAULT FALSE,
    issue_ai_actions_tip_shown BOOLEAN DEFAULT FALSE,
    project_content_tip_shown BOOLEAN DEFAULT FALSE,
    agile_board_cards_tip_shown BOOLEAN DEFAULT FALSE,
    apps_project_tab_tip_shown BOOLEAN DEFAULT FALSE,
    change_rule_type_tip_shown BOOLEAN DEFAULT FALSE,
    agile_board_visibility_tip_shown BOOLEAN DEFAULT FALSE,
    agile_board_swimlanes_tip_shown BOOLEAN DEFAULT FALSE,
    agile_board_columns_tip_shown BOOLEAN DEFAULT FALSE,
    article_ai_assistant_tip_shown BOOLEAN DEFAULT FALSE,
    article_visibility_tip_shown BOOLEAN DEFAULT FALSE,
    article_inline_comments_tip_shown BOOLEAN DEFAULT FALSE,
    project_settings_people_tip_shown BOOLEAN DEFAULT FALSE,
    project_settings_fields_tip_shown BOOLEAN DEFAULT FALSE,
    project_settings_vcs_tip_shown BOOLEAN DEFAULT FALSE,
    helpdesk_pinned_comments_tip_shown BOOLEAN DEFAULT FALSE,
    pricing_admin_popup_shown BOOLEAN DEFAULT FALSE,
    text_completion_promo_shown BOOLEAN DEFAULT FALSE,
    text_recognition_tips_shown BOOLEAN DEFAULT FALSE,
    article_sidebar_tip_shown BOOLEAN DEFAULT FALSE,
    article_comments_tip_shown BOOLEAN DEFAULT FALSE,
    activity_types_tip_shown BOOLEAN DEFAULT FALSE,
    issue_fields_tip_shown BOOLEAN DEFAULT FALSE,
    commands_tip_shown BOOLEAN DEFAULT FALSE,
    time_tracking_tip_shown BOOLEAN DEFAULT FALSE,
    visible_fields_tip_shown BOOLEAN DEFAULT FALSE,
    survey_shown BOOLEAN DEFAULT FALSE,
    votes_tip_shown BOOLEAN DEFAULT FALSE,
    ai_promo_shown BOOLEAN DEFAULT FALSE,
    collapsible_sidebar_tip_shown BOOLEAN DEFAULT FALSE,
    pmf_shown BOOLEAN DEFAULT FALSE,
    ai_writing_assistant_promo_shown BOOLEAN DEFAULT FALSE,
    inline_comment_promo_shown BOOLEAN DEFAULT TRUE,
    project_settings_time_tracking_tip_shown BOOLEAN DEFAULT FALSE,
    project_settings_teamcity_tip_shown BOOLEAN DEFAULT FALSE,
    project_settings_workflow_tip_shown BOOLEAN DEFAULT FALSE,
    project_settings_apps_tip_shown BOOLEAN DEFAULT FALSE,
    issue_text_recognition_tip_shown BOOLEAN DEFAULT FALSE,
    helpdesk_channels_tip_shown BOOLEAN DEFAULT FALSE,
    helpdesk_overview_tip_shown BOOLEAN DEFAULT FALSE,
    project_overview_page_completed BOOLEAN DEFAULT FALSE,
    project_settings_page_completed BOOLEAN DEFAULT FALSE,
    delayed_demo_modal_shown BOOLEAN DEFAULT FALSE,
    saved_searches_tip_shown BOOLEAN DEFAULT FALSE,
    search_options_tip_shown BOOLEAN DEFAULT FALSE,
    visibility_restrictions_tip_shown BOOLEAN DEFAULT FALSE,
    agile_board_backlog_tip_shown BOOLEAN DEFAULT FALSE,
    text_recognition_promo_shown BOOLEAN DEFAULT FALSE,
    ai_tips_shown BOOLEAN DEFAULT FALSE,
    agile_board_page_completed BOOLEAN DEFAULT FALSE,
    articles_page_completed BOOLEAN DEFAULT FALSE,
    helpdesk_sla_tip_shown BOOLEAN DEFAULT FALSE,
    onboarding_tour_state VARCHAR(50) DEFAULT 'idle',
    helpdesk_project_page_completed BOOLEAN DEFAULT FALSE,
    onboarding_tour_ai_block_dismissed BOOLEAN DEFAULT FALSE
);

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

CREATE TABLE IF NOT EXISTS user_profile_ai (
    user_id VARCHAR(20) PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    chats_list_show BOOLEAN DEFAULT TRUE,
    chat_floating_width INT DEFAULT 400,
    chat_mode VARCHAR(50) DEFAULT 'docked',
    chat_sidebar_show BOOLEAN DEFAULT FALSE,
    chat_floating_offset_y INT DEFAULT 100,
    chat_floating_offset_x INT DEFAULT 100,
    chat_floating_anchor VARCHAR(50) DEFAULT 'top-right',
    chat_floating_height INT DEFAULT 600,
    chat_sidebar_width INT DEFAULT 400,
    disable_chat BOOLEAN DEFAULT FALSE
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

-- -----------------------------------------------------------------------------
-- Source migration: 000005_inbox_threads.up.sql
-- -----------------------------------------------------------------------------
-- Migration to add Inbox Threads and related entities

CREATE TABLE IF NOT EXISTS inbox_threads (
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

CREATE TABLE IF NOT EXISTS inbox_messages (
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

CREATE TABLE IF NOT EXISTS inbox_message_params (
    id SERIAL PRIMARY KEY,
    message_id VARCHAR(50) REFERENCES inbox_messages(id) ON DELETE CASCADE,
    key VARCHAR(255),
    value TEXT
);

CREATE TABLE IF NOT EXISTS inbox_activities (
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
CREATE INDEX IF NOT EXISTS idx_inbox_threads_user ON inbox_threads(user_id);
CREATE INDEX IF NOT EXISTS idx_inbox_messages_thread ON inbox_messages(thread_id);
CREATE INDEX IF NOT EXISTS idx_inbox_activities_message ON inbox_activities(message_id);

-- -----------------------------------------------------------------------------
-- Source migration: 000006_banners.up.sql
-- -----------------------------------------------------------------------------
-- 000006_banners.up.sql
-- New independent schema for the banners block of GET /api/config
-- (request11.txt: fields=banners(globalBanner,globalBannerEnabled,systemEventsBanners)).
-- No existing table is modified.

CREATE TABLE IF NOT EXISTS banners_config (
    id SERIAL PRIMARY KEY,
    global_banner TEXT NOT NULL DEFAULT '',
    global_banner_enabled BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS system_events_banners (
    id SERIAL PRIMARY KEY,
    config_id INT REFERENCES banners_config(id) ON DELETE CASCADE,
    text TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_system_events_banners_config ON system_events_banners(config_id);

-- -----------------------------------------------------------------------------
-- Source migration: 000007_search_assist.up.sql
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS search_assist_responses (
    id SERIAL PRIMARY KEY,
    query TEXT,
    caret INT,
    ignore_unresolved_setting BOOLEAN DEFAULT FALSE,
    type VARCHAR(100) DEFAULT 'SearchAssistResponse'
);

CREATE TABLE IF NOT EXISTS search_style_ranges (
    id SERIAL PRIMARY KEY,
    response_id INT REFERENCES search_assist_responses(id) ON DELETE CASCADE,
    length INT,
    start_pos INT,
    style VARCHAR(100),
    title TEXT,
    type VARCHAR(100) DEFAULT 'SearchStyleRange'
);

CREATE TABLE IF NOT EXISTS search_suggestions (
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

CREATE TABLE IF NOT EXISTS search_sort_fields (
    id VARCHAR(100) PRIMARY KEY,
    name VARCHAR(255),
    sortable_presentation VARCHAR(255),
    type VARCHAR(100) DEFAULT 'SortField'
);

CREATE TABLE IF NOT EXISTS search_sort_properties (
    id SERIAL PRIMARY KEY,
    response_id INT REFERENCES search_assist_responses(id) ON DELETE CASCADE,
    property_id VARCHAR(100),
    asc_order BOOLEAN DEFAULT TRUE,
    sort_field_id VARCHAR(100) REFERENCES search_sort_fields(id),
    type VARCHAR(100) DEFAULT 'SortProperty'
);

CREATE TABLE IF NOT EXISTS search_query_features (
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

-- -----------------------------------------------------------------------------
-- Source migration: 000008_issue_count_support.up.sql
-- -----------------------------------------------------------------------------
-- request17.txt: GET /api/issuesGetter/count
-- فهارس لتسريع استعلامات العد SELECT COUNT(*) على جدول issues.

CREATE INDEX IF NOT EXISTS idx_issues_project_resolved ON issues (project_id, resolved);
CREATE INDEX IF NOT EXISTS idx_issues_updated ON issues (updated);

-- -----------------------------------------------------------------------------
-- Source migration: 000009_issue_list_subscription.up.sql
-- -----------------------------------------------------------------------------
-- 000008_issue_list_subscription.up.sql
-- New independent schema for the issue list subscription endpoint
-- (request18.txt: /api/issueListSubscription?fields=ticket).
-- No existing table is modified.

CREATE TABLE IF NOT EXISTS issue_list_subscriptions (
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

CREATE TABLE IF NOT EXISTS issue_list_subscription_issues (
    id SERIAL PRIMARY KEY,
    subscription_id INT NOT NULL REFERENCES issue_list_subscriptions(id) ON DELETE CASCADE,
    issue_id VARCHAR(50) NOT NULL,
    matches BOOLEAN NOT NULL DEFAULT TRUE,
    ordinal INT NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_issue_list_sub_issues ON issue_list_subscription_issues(subscription_id);

-- Seed a default ticket matching the request18.txt sample response.
INSERT INTO issue_list_subscriptions (ticket, query, subscribe, context_type, context_id)
VALUES ('g9qk4fe77jefpavefi1qacdhbefa3l', 'issue id: DEMO-4', TRUE, 'Project', '0-0');

-- -----------------------------------------------------------------------------
-- Source migration: 000010_global_settings_allowed_origins.up.sql
-- -----------------------------------------------------------------------------
-- 000010_global_settings_allowed_origins.up.sql
-- New independent schema to store the allowedOrigins of global settings
-- (request19.txt: /api/admin/globalSettings?fields=restSettings(allowAllOrigins,allowedOrigins),...).
-- No existing table is modified.

CREATE TABLE IF NOT EXISTS global_settings_allowed_origins (
    id SERIAL PRIMARY KEY,
    settings_id INT REFERENCES global_settings(id) ON DELETE CASCADE,
    origin VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_global_settings_allowed_origins_settings ON global_settings_allowed_origins(settings_id);

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

-- -----------------------------------------------------------------------------
-- Source migration: 000011_project_details_extended.up.sql
-- -----------------------------------------------------------------------------
-- 000011_project_details_extended.up.sql
-- Request #21: GET /api/admin/projects/{id}
-- يضيف أعمدة ناقصة إلى project_teams ومجموعة جداول لإعدادات المشروع التفصيلية
-- دون تعديل أي schema سابق.

-- Project detail tables
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

-- -----------------------------------------------------------------------------
-- Source migration: 000012_organizations_extended.up.sql
-- -----------------------------------------------------------------------------
-- 000012_organizations_extended.up.sql
-- Request #22: GET /api/admin/organizations
-- يضيف أعمدة ناقصة إلى organizations ويلقّح منظمة افتراضية مطابقة لبيانات الطلب
-- الحقيقي دون تعديل أي schema سابق (000001 إلى 000011).

-- Seed منظمة افتراضية بالمعرّف '1-0' (كما في استجابة YouTrack الحقيقية)
INSERT INTO organizations (id, key, name, icon_url, projects_count, description, audit_target_id)
VALUES ('1-0', 'CFSksvCi5N6T06bFUtA8w', 'me', NULL, 0, NULL, NULL)
ON CONFLICT (id) DO UPDATE
    SET key = EXCLUDED.key,
        name = EXCLUDED.name;

-- -----------------------------------------------------------------------------
-- Source migration: 000013_hub_services_extended.up.sql
-- -----------------------------------------------------------------------------
-- 000013_hub_services_extended.up.sql
-- Request #24: GET /hub/api/rest/services
-- يضيف أعمدة خدمة Hub الإضافية إلى جدول services دون تعديل أي schema سابق
-- (000001 إلى 000012). الأعمدة مطلوبة بواجهات Hub الأخرى في hh.json و hh2.json
-- مثل iconUrl و userUriPattern.

-- -----------------------------------------------------------------------------
-- Source migration: 000014_security_filter_fields.up.sql
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS security_filter_fields (
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

-- -----------------------------------------------------------------------------
-- Source migration: 000014_user_profile_date_field_pattern.up.sql
