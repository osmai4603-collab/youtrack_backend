-- =============================================================================
--  000007_identity_bootstrap_and_projects_seed.up.sql
--  =============================================================================
--  Purpose: Identity bootstrap (user_types, admin user, system-admin role assignment) and default project seed data.
--  Source: init_database.sql section [19-b] (lines 1744-2015).
--  Dependencies: 000001_core_schema, 000006_rbac_roles_and_auth.
-- =============================================================================

-- #############################################################################
--  [19-b] البيانات المرجعية الناقصة - BOOTSTRAP FIX
--  يجب أن يسبق هذا القسم محتوى 000019 مباشرة.
--  ليس جزءًا من الترحيلات: يملأ ثغرتين تمنعان تشغيل 000019 على قاعدة فارغة،
--  ويضيف إسناد دور الـ system-admin لحساب المدير (ثغرة لم تعالجها الترحيلات).
--  ملاحظة: هذا القسم آمن فقط لأن قسم [0] حذف كل الجداول، فلا يوجد أي صف
--  'admin' سابق يحتاج إلى مواءمة.
-- #############################################################################

-- (أ) أنواع المستخدمين المرجعية.
--     000019 يكتب user_type_id = 'standard' (الأسطر 59..61 من ملفه). طالما أن
--     user_types فارغ، يفشل أي إدراج في users على users_user_type_id_fkey.
--     القيم في YouTrack الحقيقي هي STANDARD_USER و APPLICATION_USER، لكن
--     الجداول المحلية تستعمل الرمز 'standard'، فنزرع الرموز الثلاثة معًا.
INSERT INTO user_types (id, name) VALUES
    ('standard',         'Standard User'),
    ('STANDARD_USER',    'Standard User'),
    ('APPLICATION_USER', 'Application User')
ON CONFLICT (id) DO NOTHING;

-- (ب) حساب المدير.
--     000019 ينفّذ لهذا الحساب UPDATE فقط (السطر 76) ثم يشير إليه عبر
--     projects.leader_id و issues.assignee_id. المعرّف يجب أن يبقى 'admin'
--     تحديدًا لأن البيانات المزروعة تشير بهذه القيمة.
--     000019 يحدّث email إلى osmflutterdeveloper@gmail.com، ولا يمسّ
--     password_hash، فيبقى ما نضعه هنا.
--
--     كلمة المرور الافتراضية: admin123   (bcrypt, cost 10)
INSERT INTO users (
    id, login, email, full_name, name, user_type_id,
    is_email_verified, guest, online, banned, can_read_profile, is_locked,
    password_hash
) VALUES (
    'admin', 'admin', 'admin@example.com', 'Admin', 'Admin', 'standard',
    TRUE, FALSE, FALSE, FALSE, TRUE, FALSE,
    '$2a$10$.Sa0HVENE2FSXkv9y/aTNe3PAgcsdQU9BxxE.SMhomQDF1/M0MwsO'
)
ON CONFLICT (id) DO UPDATE SET
    password_hash = EXCLUDED.password_hash;

INSERT INTO assigned_roles (id, role_id, audit_target_id, holder_type, holder_id, holder_name, scope_type, scope_id, scope_project_id, scope_organization_id)
 VALUES ('1-1', 'system-admin', NULL, 'user', 'admin', 'admin', 'global', NULL, NULL, NULL)
ON CONFLICT (id) DO NOTHING;

INSERT INTO project_types (id) VALUES ('DEFAULT')
ON CONFLICT (id) DO NOTHING;

INSERT INTO organizations (id, key, name, icon_url, projects_count, description, audit_target_id)
VALUES ('1-0', 'CFSksvCi5N6T06bFUtA8w', 'me', NULL, 1, NULL, NULL)
ON CONFLICT (id) DO UPDATE
    SET key            = EXCLUDED.key,
        name           = EXCLUDED.name,
        projects_count = EXCLUDED.projects_count;

-- ===========================================================================
-- 3. users
--    user_type_id: البيانات الحقيقية تستخدم 'STANDARD_USER' لكن الجدول المحلي
--    يحتوي 'standard' (000001_init_schema.up.sql + بذر 'user_types').
--    banned = true للحساب 'guest' وحده (حساب ضيف محظور في YouTrack) — من request51.
--    ترتيب الأعمدة: is_email_verified, guest, online, banned, can_read_profile, is_locked
-- ===========================================================================
INSERT INTO users (
    id, login, email, full_name, name, avatar_url, user_type_id,
    is_email_verified, guest, online, banned, can_read_profile, is_locked
) VALUES
('2-0', 'guest',         NULL,                     'guest',        'guest',        '/hub/api/rest/avatar/5dd5893c-e15b-4857-ba25-c37e1f959908?s=48', 'standard', FALSE, TRUE,  FALSE, TRUE,  TRUE, FALSE),
('2-3', 'osm',           'osmai4603@gmail.com',    'osm',          'osm',          '/hub/api/rest/avatar/921211c1-9f61-4784-92ad-34250abd94c0?s=48', 'standard', TRUE,  FALSE, FALSE, FALSE, TRUE, FALSE),
('2-4', 'osminventory',  'osminventory@gmail.com', 'osminventory', 'osminventory', '/hub/api/rest/avatar/0df202be-b74d-4a94-b233-2ffea5b4bbe1?s=48', 'standard', FALSE, FALSE, FALSE, FALSE, TRUE, FALSE)
ON CONFLICT (id) DO UPDATE
    SET login             = EXCLUDED.login,
        email             = EXCLUDED.email,
        full_name         = EXCLUDED.full_name,
        name              = EXCLUDED.name,
        avatar_url        = EXCLUDED.avatar_url,
        user_type_id      = EXCLUDED.user_type_id,
        is_email_verified = EXCLUDED.is_email_verified,
        guest             = EXCLUDED.guest,
        banned            = EXCLUDED.banned;

-- المستخدم leader الحقيقي ('2-1', login=admin) موجود مسبقاً بالمعرّف 'admin'.
-- يُحدَّث ملفه الشخصي ليطابق البيانات الحقيقية، مع إبقاء login و password_hash
-- كما هما (تسجيل الدخول يتم عبر عمود login فقط).
UPDATE users
SET email             = 'osmflutterdeveloper@gmail.com',
    full_name         = 'admin',
    name              = 'admin',
    avatar_url        = '/hub/api/rest/avatar/0d1219ff-dae2-4722-aa17-00381eb6be68?s=48',
    user_type_id      = 'standard',
    is_email_verified = TRUE,
    guest             = FALSE,
    banned            = FALSE
WHERE id = 'admin';

-- ===========================================================================
-- 4. user_groups
--    team_for_project_id يُترك NULL لأن projects غير مزروعة بعد
--    (user_groups.team_for_project_id <-> projects.id دائرة معكوسة).
--    القيم من request51.txt: relevantVisibilityGroups + defaultVisibilityGroup.
-- ===========================================================================
INSERT INTO user_groups (
    id, name, group_type, all_users_group, icon, description,
    audit_target_id, is_updatable, is_removable, team_for_project_id
) VALUES
('4-1',   'ahmed group',      'NestedGroup',          FALSE, NULL, NULL, '70c490e1-2496-418a-bd1c-598a33584eae', TRUE,  TRUE,  NULL),
('6-0',   'Registered Users', 'RegisteredUsersGroup', FALSE, NULL, NULL, 'e233f45d-eecf-4ea3-8994-d2d0467666a2', TRUE,  FALSE, NULL),
('103-0', 'All Users',        'AllUsersGroup',        TRUE,  NULL, NULL, '4c968b6f-11b0-4558-9b96-12713543b261', TRUE,  FALSE, NULL)
ON CONFLICT (id) DO UPDATE
    SET name             = EXCLUDED.name,
        group_type       = EXCLUDED.group_type,
        all_users_group  = EXCLUDED.all_users_group,
        audit_target_id  = EXCLUDED.audit_target_id,
        is_updatable     = EXCLUDED.is_updatable,
        is_removable     = EXCLUDED.is_removable;

-- عضوية admin في المجموعات (لمستخدمه admin بين أعضاء فرق المشروع في YouTrack).
INSERT INTO user_group_members (user_id, group_id)
VALUES
('admin', '4-1'),
('admin', '6-0'),
('admin', '103-0')
ON CONFLICT (user_id, group_id) DO NOTHING;

-- ===========================================================================
-- 5. projects
--    team_id = NULL هنا لكسر الدائرة مع project_teams (تُملأ في الخطوة 7).
--    القيم الكاملة للصف 0-0 مأخوذة من request51.txt.
--    الصفوف 0-2..0-4: short_name مشتق (غير موجود في الالتقاطات) و
--    creation_time = NULL (غير معروف).
--    default_visibility_group_id = '4-1' لمشروع 0-0 (defaultVisibilityGroup).
-- ===========================================================================
INSERT INTO projects (
    id, name, short_name, project_type_id, pinned, icon_url, template, archived,
    restricted, has_articles, is_demo, fields_sorted, query, issues_url,
    description, leader_id, team_id, organization_id, creation_time,
    from_email, from_personal, reply_to_email, email_delimiter,
    use_email_delimiter, supports_email_delimiter, default_visibility_group_id,
    default_smtp, source_template, audit_target_id, historical_short_names,
    created_at, updated_at
) VALUES
('0-0', 'Demo project',         'DEMO',   'DEFAULT', TRUE,  NULL, FALSE, FALSE, FALSE, TRUE,  TRUE,  FALSE,
 'project: {Demo project}',          '/issues/DEMO',   NULL, 'admin', NULL, '1-0', 1784851998350,
 'osmai', 'osmai', 'osmflutterdeveloper@gmail.com',
 'Please enter your reply above this line', FALSE, FALSE, '4-1',
 TRUE,  NULL, '0-0', ARRAY[]::TEXT[], 1784851998350, 1784851998350),
('0-1', 'fingerprint',          'FIN',    'DEFAULT', TRUE,  NULL, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE,
 'project: {fingerprint}',           '/issues/FIN',    NULL, NULL,   NULL, NULL,  NULL,
 NULL, NULL, NULL, NULL, FALSE, FALSE, NULL,
 FALSE, NULL, '0-1', ARRAY[]::TEXT[], 1784851998350, 1784851998350),
('0-2', 'Test project',         'TEST',   'DEFAULT', FALSE, NULL, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE,
 'project: {Test project}',          '/issues/TEST',   NULL, NULL,   NULL, NULL,  NULL,
 NULL, NULL, NULL, NULL, FALSE, FALSE, NULL,
 FALSE, NULL, '0-2', ARRAY[]::TEXT[], 1784851998350, 1784851998350),
('0-3', 'pringo',               'PRINGO', 'DEFAULT', FALSE, NULL, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE,
 'project: {pringo}',                '/issues/PRINGO', NULL, NULL,   NULL, NULL,  NULL,
 NULL, NULL, NULL, NULL, FALSE, FALSE, NULL,
 FALSE, NULL, '0-3', ARRAY[]::TEXT[], 1784851998350, 1784851998350),
('0-4', 'traint-project-scrum', 'TRAINT', 'DEFAULT', FALSE, NULL, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE,
 'project: {traint-project-scrum}',  '/issues/TRAINT', NULL, NULL,  NULL, NULL,  NULL,
 NULL, NULL, NULL, NULL, FALSE, FALSE, NULL,
 FALSE, NULL, '0-4', ARRAY[]::TEXT[], 1784851998350, 1784851998350)
ON CONFLICT (id) DO UPDATE
    SET name                       = EXCLUDED.name,
        short_name                 = EXCLUDED.short_name,
        project_type_id            = EXCLUDED.project_type_id,
        pinned                     = EXCLUDED.pinned,
        icon_url                   = EXCLUDED.icon_url,
        template                   = EXCLUDED.template,
        archived                   = EXCLUDED.archived,
        restricted                 = EXCLUDED.restricted,
        has_articles               = EXCLUDED.has_articles,
        is_demo                    = EXCLUDED.is_demo,
        fields_sorted              = EXCLUDED.fields_sorted,
        query                      = EXCLUDED.query,
        issues_url                 = EXCLUDED.issues_url,
        description                = EXCLUDED.description,
        leader_id                  = EXCLUDED.leader_id,
        organization_id            = EXCLUDED.organization_id,
        creation_time              = EXCLUDED.creation_time,
        from_email                 = EXCLUDED.from_email,
        from_personal              = EXCLUDED.from_personal,
        reply_to_email             = EXCLUDED.reply_to_email,
        email_delimiter            = EXCLUDED.email_delimiter,
        use_email_delimiter        = EXCLUDED.use_email_delimiter,
        supports_email_delimiter   = EXCLUDED.supports_email_delimiter,
        default_visibility_group_id = EXCLUDED.default_visibility_group_id,
        default_smtp               = EXCLUDED.default_smtp,
        source_template            = EXCLUDED.source_template,
        audit_target_id            = EXCLUDED.audit_target_id,
        historical_short_names     = EXCLUDED.historical_short_names,
        created_at                 = EXCLUDED.created_at,
        updated_at                 = EXCLUDED.updated_at;

-- ===========================================================================
-- 6. project_teams
--    أسماء الفرق من request51.txt ("Demo project Team") و request62.txt.
--    team_id في projects يُملأ في الخطوة 7.
-- ===========================================================================
INSERT INTO project_teams (id, name, project_id) VALUES
('5-0', 'Demo project Team',         '0-0'),
('5-1', 'fingerprint Team',          '0-1'),
('5-2', 'Test project Team',         '0-2'),
('5-3', 'pringo Team',               '0-3'),
('5-4', 'traint-project-scrum Team', '0-4')
ON CONFLICT (id) DO UPDATE
    SET name       = EXCLUDED.name,
        project_id = EXCLUDED.project_id;

-- عضوية الفرق (project_team_members) من request27.txt — فريق مشروع DEMO:
-- guest (2-0), admin (المعرّف المحلي 'admin'), osm (2-3), osminventory (2-4).
INSERT INTO project_team_members (team_id, user_id) VALUES
('5-0', '2-0'),
('5-0', 'admin'),
('5-0', '2-3'),
('5-0', '2-4')
ON CONFLICT (team_id, user_id) DO NOTHING;

-- ===========================================================================
-- 7. ربط projects.team_id بعد اكتمال project_teams (كسر الدائرة)
-- ===========================================================================
UPDATE projects SET team_id = '5-0' WHERE id = '0-0';
UPDATE projects SET team_id = '5-1' WHERE id = '0-1';
UPDATE projects SET team_id = '5-2' WHERE id = '0-2';
UPDATE projects SET team_id = '5-3' WHERE id = '0-3';
UPDATE projects SET team_id = '5-4' WHERE id = '0-4';


