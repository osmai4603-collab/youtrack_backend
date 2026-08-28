-- Drop existing tables to ensure clean state
DROP TABLE IF EXISTS assigned_roles CASCADE;
DROP TABLE IF EXISTS permission_dependencies CASCADE;
DROP TABLE IF EXISTS role_permissions CASCADE;
DROP TABLE IF EXISTS roles CASCADE;
DROP TABLE IF EXISTS cached_permission_projects CASCADE;
DROP TABLE IF EXISTS cached_permissions CASCADE;
DROP TABLE IF EXISTS permissions CASCADE;
DROP TABLE IF EXISTS recent_articles CASCADE;
DROP TABLE IF EXISTS recent_issues CASCADE;
DROP TABLE IF EXISTS visibility_permitted_groups CASCADE;
DROP TABLE IF EXISTS visibility_permitted_users CASCADE;
DROP TABLE IF EXISTS visibility CASCADE;
DROP TABLE IF EXISTS issue_watchers CASCADE;
DROP TABLE IF EXISTS issue_voters CASCADE;
DROP TABLE IF EXISTS issue_tags CASCADE;
DROP TABLE IF EXISTS tags CASCADE;
DROP TABLE IF EXISTS issue_custom_field_values CASCADE;
DROP TABLE IF EXISTS issue_field_enum_values CASCADE;
DROP TABLE IF EXISTS issue_links CASCADE;
DROP TABLE IF EXISTS comments CASCADE;
DROP TABLE IF EXISTS attachments CASCADE;
DROP TABLE IF EXISTS work_items CASCADE;
DROP TABLE IF EXISTS issues CASCADE;
DROP TABLE IF EXISTS articles CASCADE;
DROP TABLE IF EXISTS project_team_members CASCADE;
DROP TABLE IF EXISTS project_teams CASCADE;
DROP TABLE IF EXISTS user_group_members CASCADE;
DROP TABLE IF EXISTS user_groups CASCADE;
DROP TABLE IF EXISTS user_profiles CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS user_types CASCADE;
DROP TABLE IF EXISTS project_app_configurations CASCADE;
DROP TABLE IF EXISTS app_tags CASCADE;
DROP TABLE IF EXISTS apps CASCADE;
DROP TABLE IF EXISTS project_dashboard_widgets CASCADE;
DROP TABLE IF EXISTS dashboard_widgets CASCADE;
DROP TABLE IF EXISTS project_time_tracking_settings CASCADE;
DROP TABLE IF EXISTS work_item_types CASCADE;
DROP TABLE IF EXISTS work_days CASCADE;
DROP TABLE IF EXISTS work_time_settings CASCADE;
DROP TABLE IF EXISTS board_cells CASCADE;
DROP TABLE IF EXISTS board_column_field_values CASCADE;
DROP TABLE IF EXISTS board_columns CASCADE;
DROP TABLE IF EXISTS sprints CASCADE;
DROP TABLE IF EXISTS agile_board_projects CASCADE;
DROP TABLE IF EXISTS swimlane_settings CASCADE;
DROP TABLE IF EXISTS agile_boards CASCADE;
DROP TABLE IF EXISTS saved_queries CASCADE;
DROP TABLE IF EXISTS project_custom_fields CASCADE;
DROP TABLE IF EXISTS custom_fields CASCADE;
DROP TABLE IF EXISTS field_converters CASCADE;
DROP TABLE IF EXISTS field_types CASCADE;
DROP TABLE IF EXISTS bundle_values CASCADE;
DROP TABLE IF EXISTS bundles CASCADE;
DROP TABLE IF EXISTS field_styles CASCADE;
DROP TABLE IF EXISTS projects CASCADE;
DROP TABLE IF EXISTS project_types CASCADE;
DROP TABLE IF EXISTS organizations CASCADE;
DROP TABLE IF EXISTS services CASCADE;
DROP TABLE IF EXISTS vcs_hosting_servers CASCADE;
DROP TABLE IF EXISTS feature_flags CASCADE;
DROP TABLE IF EXISTS inbox_folders CASCADE;
DROP TABLE IF EXISTS global_settings CASCADE;
DROP TABLE IF EXISTS filter_fields CASCADE;
DROP TABLE IF EXISTS notification_templates CASCADE;
DROP TABLE IF EXISTS notification_template_groups CASCADE;

-- 1. Independent & Base Tables
CREATE TABLE field_styles (
    id VARCHAR(20) PRIMARY KEY,
    background VARCHAR(20),
    foreground VARCHAR(20)
);

CREATE TABLE project_types (
    id VARCHAR(50) PRIMARY KEY
);

CREATE TABLE user_types (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL
);

CREATE TABLE organizations (
    id VARCHAR(20) PRIMARY KEY,
    key VARCHAR(100),
    name VARCHAR(255),
    icon_url TEXT,
    projects_count INT DEFAULT 0
);

CREATE TABLE permissions (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255),
    description TEXT,
    permission_entity_type VARCHAR(100),
    localized_permission_entity_type VARCHAR(100),
    operation VARCHAR(50),
    is_global BOOLEAN DEFAULT FALSE
);

CREATE TABLE work_time_settings (
    id SERIAL PRIMARY KEY,
    minutes_a_day INT DEFAULT 480,
    minutes_a_day_presentation VARCHAR(10) DEFAULT '8h',
    days_a_week INT DEFAULT 5,
    first_day_of_week INT DEFAULT 0
);

CREATE TABLE work_days (
    settings_id INT REFERENCES work_time_settings(id) ON DELETE CASCADE,
    day_number INT,
    PRIMARY KEY (settings_id, day_number)
);

CREATE TABLE work_item_types (
    id VARCHAR(20) PRIMARY KEY,
    name VARCHAR(255),
    color_id VARCHAR(20) REFERENCES field_styles(id) ON DELETE SET NULL,
    auto_attach BOOLEAN DEFAULT FALSE,
    description TEXT
);

CREATE TABLE services (
    id VARCHAR(100) PRIMARY KEY,
    name VARCHAR(255),
    key VARCHAR(255),
    home_url TEXT,
    application_name VARCHAR(255),
    vendor VARCHAR(255),
    version VARCHAR(50),
    trusted BOOLEAN DEFAULT FALSE
);

CREATE TABLE vcs_hosting_servers (
    id VARCHAR(20) PRIMARY KEY,
    url TEXT,
    server_type VARCHAR(50),
    is_predefined BOOLEAN DEFAULT TRUE,
    app_id VARCHAR(100),
    app_name VARCHAR(255),
    application_id VARCHAR(100),
    ssl_key_id VARCHAR(20)
);

CREATE TABLE feature_flags (
    id VARCHAR(255) PRIMARY KEY,
    enabled BOOLEAN DEFAULT FALSE
);

CREATE TABLE field_types (
    id VARCHAR(50) PRIMARY KEY,
    presentation VARCHAR(100),
    value_type VARCHAR(50),
    is_bundle_type BOOLEAN DEFAULT FALSE,
    is_multi_value BOOLEAN DEFAULT FALSE
);

CREATE TABLE field_converters (
    id VARCHAR(100) PRIMARY KEY,
    from_type_id VARCHAR(50) REFERENCES field_types(id) ON DELETE CASCADE,
    to_type_id VARCHAR(50) REFERENCES field_types(id) ON DELETE CASCADE,
    localized_name VARCHAR(255)
);

CREATE TABLE bundles (
    id VARCHAR(20) PRIMARY KEY,
    bundle_type VARCHAR(50) NOT NULL,
    name VARCHAR(255),
    is_updateable BOOLEAN DEFAULT FALSE
);

-- 2. Core Entities with FKs
CREATE TABLE users (
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
    can_read_profile BOOLEAN DEFAULT TRUE,
    is_locked BOOLEAN DEFAULT FALSE,
    ring_id VARCHAR(100)
);

CREATE TABLE user_profiles (
    user_id VARCHAR(20) PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    timezone_id VARCHAR(100),
    locale_id VARCHAR(20),
    date_pattern VARCHAR(50),
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

CREATE TABLE user_groups (
    id VARCHAR(20) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    group_type VARCHAR(50),
    all_users_group BOOLEAN DEFAULT FALSE,
    icon TEXT,
    description TEXT,
    audit_target_id VARCHAR(100),
    is_updatable BOOLEAN DEFAULT FALSE,
    is_removable BOOLEAN DEFAULT FALSE,
    team_for_project_id VARCHAR(20)
);

CREATE TABLE user_group_members (
    user_id VARCHAR(20) REFERENCES users(id) ON DELETE CASCADE,
    group_id VARCHAR(20) REFERENCES user_groups(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, group_id)
);

CREATE TABLE projects (
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
    team_id VARCHAR(20),
    organization_id VARCHAR(20) REFERENCES organizations(id) ON DELETE SET NULL,
    creation_time BIGINT,
    from_email VARCHAR(255),
    from_personal VARCHAR(255),
    reply_to_email VARCHAR(255),
    email_delimiter TEXT,
    use_email_delimiter BOOLEAN DEFAULT FALSE,
    supports_email_delimiter BOOLEAN DEFAULT FALSE,
    default_visibility_group_id VARCHAR(20) REFERENCES user_groups(id) ON DELETE SET NULL,
    default_smtp BOOLEAN DEFAULT FALSE,
    source_template VARCHAR(20),
    audit_target_id VARCHAR(100),
    historical_short_names TEXT[],
    created_at BIGINT,
    updated_at BIGINT
);

CREATE TABLE project_teams (
    id VARCHAR(20) PRIMARY KEY,
    name VARCHAR(255),
    project_id VARCHAR(20) REFERENCES projects(id) ON DELETE CASCADE
);

ALTER TABLE projects ADD CONSTRAINT fk_projects_team FOREIGN KEY (team_id) REFERENCES project_teams(id) ON DELETE SET NULL;
ALTER TABLE user_groups ADD CONSTRAINT fk_user_groups_team_project FOREIGN KEY (team_for_project_id) REFERENCES projects(id) ON DELETE SET NULL;

CREATE TABLE project_team_members (
    team_id VARCHAR(20) REFERENCES project_teams(id) ON DELETE CASCADE,
    user_id VARCHAR(20) REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (team_id, user_id)
);

-- 3. Custom Fields & Bundles
CREATE TABLE custom_fields (
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

CREATE TABLE project_custom_fields (
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

CREATE TABLE bundle_values (
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
CREATE TABLE saved_queries (
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

CREATE TABLE issues (
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

CREATE TABLE comments (
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

CREATE TABLE articles (
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

CREATE TABLE attachments (
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

CREATE TABLE work_items (
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

CREATE TABLE issue_links (
    id VARCHAR(20) PRIMARY KEY,
    source_issue_id VARCHAR(20) NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    target_issue_id VARCHAR(20) NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    link_type VARCHAR(100),
    direction VARCHAR(20)
);

CREATE TABLE tags (
    id VARCHAR(20) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    color_id VARCHAR(20) REFERENCES field_styles(id) ON DELETE SET NULL,
    is_deletable BOOLEAN DEFAULT FALSE,
    is_updatable BOOLEAN DEFAULT FALSE,
    is_usable BOOLEAN DEFAULT FALSE
);

CREATE TABLE issue_tags (
    issue_id VARCHAR(20) REFERENCES issues(id) ON DELETE CASCADE,
    tag_id VARCHAR(20) REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (issue_id, tag_id)
);

CREATE TABLE issue_custom_field_values (
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

CREATE TABLE issue_field_enum_values (
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

CREATE TABLE issue_voters (
    issue_id VARCHAR(20) REFERENCES issues(id) ON DELETE CASCADE,
    user_id VARCHAR(20) REFERENCES users(id) ON DELETE CASCADE,
    has_vote BOOLEAN DEFAULT FALSE,
    PRIMARY KEY (issue_id, user_id)
);

CREATE TABLE issue_watchers (
    issue_id VARCHAR(20) REFERENCES issues(id) ON DELETE CASCADE,
    user_id VARCHAR(20) REFERENCES users(id) ON DELETE CASCADE,
    has_star BOOLEAN DEFAULT FALSE,
    PRIMARY KEY (issue_id, user_id)
);

-- 5. Visibility
CREATE TABLE visibility (
    id VARCHAR(20) PRIMARY KEY,
    visibility_type VARCHAR(50),
    entity_type VARCHAR(50),
    entity_id VARCHAR(20)
);

CREATE TABLE visibility_permitted_users (
    visibility_id VARCHAR(20) REFERENCES visibility(id) ON DELETE CASCADE,
    user_id VARCHAR(20) REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (visibility_id, user_id)
);

CREATE TABLE visibility_permitted_groups (
    visibility_id VARCHAR(20) REFERENCES visibility(id) ON DELETE CASCADE,
    group_id VARCHAR(20) REFERENCES user_groups(id) ON DELETE CASCADE,
    PRIMARY KEY (visibility_id, group_id)
);

-- 6. Recents & Inbox
CREATE TABLE recent_issues (
    id VARCHAR(20) PRIMARY KEY,
    user_id VARCHAR(20) REFERENCES users(id) ON DELETE CASCADE,
    issue_id VARCHAR(20) REFERENCES issues(id) ON DELETE CASCADE,
    pinned BOOLEAN DEFAULT FALSE,
    date BIGINT
);

CREATE TABLE recent_articles (
    id VARCHAR(20) PRIMARY KEY,
    user_id VARCHAR(20) REFERENCES users(id) ON DELETE CASCADE,
    article_id VARCHAR(20) REFERENCES articles(id) ON DELETE CASCADE,
    pinned BOOLEAN DEFAULT FALSE,
    date BIGINT
);

CREATE TABLE inbox_folders (
    id VARCHAR(50) PRIMARY KEY,
    user_id VARCHAR(20) REFERENCES users(id) ON DELETE CASCADE,
    last_seen BIGINT DEFAULT 0,
    last_notified BIGINT DEFAULT 0,
    enabled BOOLEAN DEFAULT TRUE
);

-- 7. Roles & Permissions Cache
CREATE TABLE roles (
    id VARCHAR(100) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    audit_target_id VARCHAR(100),
    is_updatable BOOLEAN DEFAULT FALSE,
    immutable BOOLEAN DEFAULT FALSE
);

CREATE TABLE role_permissions (
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

CREATE TABLE permission_dependencies (
    permission_id VARCHAR(255) REFERENCES permissions(id) ON DELETE CASCADE,
    dependent_permission_id VARCHAR(255) REFERENCES permissions(id) ON DELETE CASCADE,
    dependent_permission_name VARCHAR(255),
    PRIMARY KEY (permission_id, dependent_permission_id)
);

CREATE TABLE assigned_roles (
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

CREATE TABLE cached_permissions (
    id VARCHAR(100) REFERENCES permissions(id) ON DELETE CASCADE,
    user_id VARCHAR(20) REFERENCES users(id) ON DELETE CASCADE,
    is_global BOOLEAN DEFAULT FALSE,
    PRIMARY KEY (id, user_id)
);

CREATE TABLE cached_permission_projects (
    permission_id VARCHAR(100) REFERENCES permissions(id) ON DELETE CASCADE,
    project_id VARCHAR(20) REFERENCES projects(id) ON DELETE CASCADE,
    PRIMARY KEY (permission_id, project_id)
);

-- 8. Agile & Boards
CREATE TABLE agile_boards (
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

CREATE TABLE agile_board_projects (
    agile_id VARCHAR(20) REFERENCES agile_boards(id) ON DELETE CASCADE,
    project_id VARCHAR(20) REFERENCES projects(id) ON DELETE CASCADE,
    PRIMARY KEY (agile_id, project_id)
);

CREATE TABLE sprints (
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

CREATE TABLE board_columns (
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

CREATE TABLE board_column_field_values (
    id VARCHAR(20) PRIMARY KEY,
    column_id VARCHAR(20) REFERENCES board_columns(id) ON DELETE CASCADE,
    name VARCHAR(255),
    presentation VARCHAR(255),
    ordinal INT DEFAULT 0,
    is_resolved BOOLEAN DEFAULT FALSE,
    can_update BOOLEAN DEFAULT FALSE
);

CREATE TABLE board_cells (
    id VARCHAR(50) PRIMARY KEY,
    column_id VARCHAR(20) REFERENCES board_columns(id) ON DELETE CASCADE,
    row_id VARCHAR(50),
    issues_count INT DEFAULT 0,
    too_many_issues BOOLEAN DEFAULT FALSE
);

CREATE TABLE swimlane_settings (
    id VARCHAR(20) PRIMARY KEY,
    agile_id VARCHAR(20) REFERENCES agile_boards(id) ON DELETE CASCADE,
    enabled BOOLEAN DEFAULT FALSE,
    field_id VARCHAR(20) REFERENCES custom_fields(id) ON DELETE SET NULL,
    default_card_type VARCHAR(50)
);

-- 9. Apps & Dashboards
CREATE TABLE apps (
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

CREATE TABLE app_tags (
    app_id VARCHAR(20) REFERENCES apps(id) ON DELETE CASCADE,
    name VARCHAR(100),
    PRIMARY KEY (app_id, name)
);

CREATE TABLE project_app_configurations (
    id VARCHAR(20) PRIMARY KEY,
    project_id VARCHAR(20) REFERENCES projects(id) ON DELETE CASCADE,
    app_id VARCHAR(20) REFERENCES apps(id) ON DELETE CASCADE,
    enabled BOOLEAN DEFAULT TRUE,
    is_broken BOOLEAN DEFAULT FALSE,
    can_update BOOLEAN DEFAULT FALSE,
    missing_required_settings BOOLEAN DEFAULT FALSE,
    project_settings TEXT
);

CREATE TABLE dashboard_widgets (
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
    guard TEXT
);

CREATE TABLE project_dashboard_widgets (
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

CREATE TABLE project_time_tracking_settings (
    id VARCHAR(20) PRIMARY KEY,
    project_id VARCHAR(20) REFERENCES projects(id) ON DELETE CASCADE,
    enabled BOOLEAN DEFAULT FALSE,
    estimate_field_id VARCHAR(20) REFERENCES project_custom_fields(id) ON DELETE SET NULL,
    time_spent_field_id VARCHAR(20) REFERENCES project_custom_fields(id) ON DELETE SET NULL
);

-- 10. Global & Notification Settings
CREATE TABLE global_settings (
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

CREATE TABLE filter_fields (
    id VARCHAR(100) PRIMARY KEY,
    name VARCHAR(255),
    field_type VARCHAR(50),
    custom_field_id VARCHAR(20) REFERENCES custom_fields(id) ON DELETE SET NULL
);

CREATE TABLE notification_templates (
    id VARCHAR(255) PRIMARY KEY,
    file_name VARCHAR(255),
    description TEXT,
    content TEXT NOT NULL,
    overrided BOOLEAN DEFAULT FALSE
);

CREATE TABLE notification_template_groups (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255),
    title VARCHAR(255),
    description TEXT,
    enabled BOOLEAN DEFAULT TRUE,
    customizable BOOLEAN DEFAULT FALSE,
    parent_group_id VARCHAR(255) REFERENCES notification_template_groups(id) ON DELETE SET NULL
);
