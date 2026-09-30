-- 000015_seed_test_data.up.sql
-- بذر بيانات اختبار واقعية لتطبيق صلاحيات الوصول على مستوى المشاريع.
--
-- ثلاثة مستخدمين (test1/test2/test3) بكلمة مرور admin123 (مخزّنة كـ bcrypt hash).
--   * test1 يقود مشروعين: proj-t1-a (T1A), proj-t1-b (T1B)
--   * test2 يقود مشروعين: proj-t2-a (T2A), proj-t2-b (T2B)
--   * test3 لا يملك أي مشروع (يجب أن يرى قوائم فارغة)
-- لكل مشروع 3 قضايا => 12 قضية إجمالاً.
--
-- ملاحظات التنفيذ:
--   * الملف يُنفَّذ داخل معاملة واحدة (Tx) من مُشغّل الـ migrations
--     (انظر channels/store/sqlstore/migrate.go) فلا حاجة لـ BEGIN/COMMIT هنا،
--     وأي خطأ يؤدي إلى تراجع كامل.
--   * كل إدراج يستخدم ON CONFLICT DO NOTHING ليبقى الملف قابلاً لإعادة التشغيل.
--   * ترتيب الإدراج يحترم المفاتيح الأجنبية.(projects.team_id ⇄ project_teams.project_id)
--     علاقة دائرية، لذلك تُنشأ الفرق أولاً بدون project_id ثم تُربط بعد إنشاء المشاريع.
--   * كلمات المرور لا تُخزَّن أبداً كنص صريح، بل كـ bcrypt hash فقط.

-- 1. أنواع المستخدمين وأنواع المشاريع (أساس FK للمستخدمين والمشاريع)
INSERT INTO user_types (id, name) VALUES
    ('STANDARD_USER', 'Standard user')
ON CONFLICT (id) DO NOTHING;

INSERT INTO project_types (id) VALUES
    ('DEFAULT')
ON CONFLICT (id) DO NOTHING;

-- 2. المستخدمون: معرّف المستخدم = اسم الدخول (يطابق claim "sub" في رمز JWT)
--    كلمة المرور：admin123
--    ملاحظة مهمة: أعمدة النصوص (avatar_url, ring_id ...) تُكتب بقيمة فارغة '' بدل NULL،
--    لأن دالة القراءة في channels/store/sqlstore/user_store.go:scanUser تفشل عند مسح NULL
--    في خانة من نوع string (فتعشل المصادقة). يُنصح بإصلاح ذلك في المخزن لاحقاً.
INSERT INTO users (id, login, email, full_name, name, avatar_url, ring_id, user_type_id,
                   is_email_verified, guest, online, banned, can_read_profile, is_locked, password_hash)
VALUES
    ('test1', 'test1', 'test1@youtrack.local', 'Test One',   'test1',
     '/hub/api/rest/avatar/t1a00000-0000-0000-0000-000000000001?s=48', '', 'STANDARD_USER',
     TRUE, FALSE, TRUE, FALSE, TRUE, FALSE, '$2a$10$fZRjMe.6mx8/9x7zSqwNeed8j.zpvqWmSjVOpMK18OkqdxqKNnOBC'),
    ('test2', 'test2', 'test2@youtrack.local', 'Test Two',   'test2',
     '/hub/api/rest/avatar/t2b00000-0000-0000-0000-000000000002?s=48', '', 'STANDARD_USER',
     TRUE, FALSE, TRUE, FALSE, TRUE, FALSE, '$2a$10$fZRjMe.6mx8/9x7zSqwNeed8j.zpvqWmSjVOpMK18OkqdxqKNnOBC'),
    ('test3', 'test3', 'test3@youtrack.local', 'Test Three', 'test3',
     '/hub/api/rest/avatar/t3c00000-0000-0000-0000-000000000003?s=48', '', 'STANDARD_USER',
     TRUE, FALSE, TRUE, FALSE, TRUE, FALSE, '$2a$10$fZRjMe.6mx8/9x7zSqwNeed8j.zpvqWmSjVOpMK18OkqdxqKNnOBC')
ON CONFLICT (id) DO NOTHING;

-- 3. ملفات تعريف المستخدمين
INSERT INTO user_profiles (user_id, timezone_id, locale_id, date_pattern,
                           email_notifications_enabled, mention_notifications_enabled,
                           auto_watch_on_create, auto_watch_on_comment, compact_mode,
                           expand_navigation, natural_comments_order, use_markdown_editor,
                           show_sidebar, unresolved_issues_only, is_time_tracking_available)
VALUES
    ('test1', 'Etc/UTC', 'en', 'yyyy-MM-dd', TRUE, TRUE, TRUE, TRUE, FALSE, TRUE, TRUE, FALSE, TRUE, FALSE, FALSE),
    ('test2', 'Etc/UTC', 'en', 'yyyy-MM-dd', TRUE, TRUE, TRUE, TRUE, FALSE, TRUE, TRUE, FALSE, TRUE, FALSE, FALSE),
    ('test3', 'Etc/UTC', 'en', 'yyyy-MM-dd', TRUE, TRUE, TRUE, TRUE, FALSE, TRUE, TRUE, FALSE, TRUE, FALSE, FALSE)
ON CONFLICT (user_id) DO NOTHING;

-- 4. فرق المشاريع (تُنشأ قبل المشاريع لأن projects.team_id يشير إليها، وتُربط بـ project_id بعد إنشاء المشاريع في الخطوة 6).
INSERT INTO project_teams (id, name) VALUES
    ('team-t1-a', 'T1A Project Team'),
    ('team-t1-b', 'T1B Project Team'),
    ('team-t2-a', 'T2A Project Team'),
    ('team-t2-b', 'T2B Project Team')
ON CONFLICT (id) DO NOTHING;

-- 5. المشاريع: يقوده المستخدم (leader_id) + له فريق (team_id)
--    ملاحظة: كل أعمدة النصوص والأرقام المنطقية تُملأ بقيمة صريحة (وليس NULL) لأن
--    دوال القراءة في channels/store/sqlstore/project_store.go تفشل عند مسح NULL.
INSERT INTO projects (id, name, short_name, project_type_id, pinned, icon_url, template, archived,
                      restricted, has_articles, is_demo, fields_sorted, query, issues_url, description,
                      leader_id, team_id, organization_id, creation_time, default_smtp, source_template,
                      audit_target_id, historical_short_names, from_email, from_personal, reply_to_email,
                      email_delimiter, supports_email_delimiter, use_email_delimiter,
                      created_at, updated_at)
VALUES
    ('proj-t1-a', 'Test1 Project Alpha', 'T1A', 'DEFAULT', TRUE, '', FALSE, FALSE,
     FALSE, FALSE, FALSE, FALSE, 'project: T1A', '', 'Seed project owned by test1',
     'test1', 'team-t1-a', '1-0', 1700000000000, FALSE, '', 'proj-t1-a', '{}', '', '', '',
     '', FALSE, FALSE, 1700000000000, 1700000000000),
    ('proj-t1-b', 'Test1 Project Beta',  'T1B', 'DEFAULT', FALSE, '', FALSE, FALSE,
     FALSE, FALSE, FALSE, FALSE, 'project: T1B', '', 'Seed project owned by test1',
     'test1', 'team-t1-b', '1-0', 1700000000000, FALSE, '', 'proj-t1-b', '{}', '', '', '',
     '', FALSE, FALSE, 1700000000000, 1700000000000),
    ('proj-t2-a', 'Test2 Project Alpha', 'T2A', 'DEFAULT', TRUE, '', FALSE, FALSE,
     FALSE, FALSE, FALSE, FALSE, 'project: T2A', '', 'Seed project owned by test2',
     'test2', 'team-t2-a', '1-0', 1700000000000, FALSE, '', 'proj-t2-a', '{}', '', '', '',
     '', FALSE, FALSE, 1700000000000, 1700000000000),
    ('proj-t2-b', 'Test2 Project Beta',  'T2B', 'DEFAULT', FALSE, '', FALSE, FALSE,
     FALSE, FALSE, FALSE, FALSE, 'project: T2B', '', 'Seed project owned by test2',
     'test2', 'team-t2-b', '1-0', 1700000000000, FALSE, '', 'proj-t2-b', '{}', '', '', '',
     '', FALSE, FALSE, 1700000000000, 1700000000000)
ON CONFLICT (id) DO NOTHING;

-- 6. ربط الفرق بالمشاريع (استكمال العلاقة الدائرية مع الخطوة 5)
UPDATE project_teams SET project_id = 'proj-t1-a' WHERE id = 'team-t1-a';
UPDATE project_teams SET project_id = 'proj-t1-b' WHERE id = 'team-t1-b';
UPDATE project_teams SET project_id = 'proj-t2-a' WHERE id = 'team-t2-a';
UPDATE project_teams SET project_id = 'proj-t2-b' WHERE id = 'team-t2-b';

-- 7. أعضاء فرق المشاريع: كل مستخدم عضو في فريق مشاريعه
INSERT INTO project_team_members (team_id, user_id) VALUES
    ('team-t1-a', 'test1'),
    ('team-t1-b', 'test1'),
    ('team-t2-a', 'test2'),
    ('team-t2-b', 'test2')
ON CONFLICT (team_id, user_id) DO NOTHING;

-- 8. مجموعات المستخدمين: مجموعة افتراضية لكل مشروع (تُستخدم كـ defaultVisibilityGroup)
INSERT INTO user_groups (id, name, group_type, all_users_group, description,
                         is_updatable, is_removable, team_for_project_id)
VALUES
    ('grp-t1-a', 'T1A Project Members',  'PROJECT', FALSE, 'Members of T1A', TRUE, TRUE, 'proj-t1-a'),
    ('grp-t1-b', 'T1B Project Members',  'PROJECT', FALSE, 'Members of T1B', TRUE, TRUE, 'proj-t1-b'),
    ('grp-t2-a', 'T2A Project Members',  'PROJECT', FALSE, 'Members of T2A', TRUE, TRUE, 'proj-t2-a'),
    ('grp-t2-b', 'T2B Project Members',  'PROJECT', FALSE, 'Members of T2B', TRUE, TRUE, 'proj-t2-b')
ON CONFLICT (id) DO NOTHING;

INSERT INTO user_group_members (user_id, group_id) VALUES
    ('test1', 'grp-t1-a'),
    ('test1', 'grp-t1-b'),
    ('test2', 'grp-t2-a'),
    ('test2', 'grp-t2-b')
ON CONFLICT (user_id, group_id) DO NOTHING;

-- 9. جعل مجموعة المشروع هي مجموعة الرؤية الافتراضية
UPDATE projects SET default_visibility_group_id = 'grp-t1-a' WHERE id = 'proj-t1-a';
UPDATE projects SET default_visibility_group_id = 'grp-t1-b' WHERE id = 'proj-t1-b';
UPDATE projects SET default_visibility_group_id = 'grp-t2-a' WHERE id = 'proj-t2-a';
UPDATE projects SET default_visibility_group_id = 'grp-t2-b' WHERE id = 'proj-t2-b';

-- 10. القضايا: 3 قضايا لكل مشروع (12 قضية). التوقيت بالمللي ثانية، تصاعدي.
--     resolved = 0 تعني "غير محلولة" (تُملأ صراحةً لتفادي فشل مسح NULL في المخزن).
INSERT INTO issues (id, id_readable, number_in_project, summary, description, project_id,
                    reporter_id, creator_id, updater_id, created, updated, resolved, votes, is_draft)
VALUES
    ('iss-t1a-1', 'T1A-1', 1, 'Setup CI pipeline',    'Configure CI for T1A',       'proj-t1-a', 'test1', 'test1', 'test1', 1700000001000, 1700000001000, 0, 0, FALSE),
    ('iss-t1a-2', 'T1A-2', 2, 'Add unit tests',       'Cover services with tests',  'proj-t1-a', 'test1', 'test1', 'test1', 1700000002000, 1700000002000, 0, 0, FALSE),
    ('iss-t1a-3', 'T1A-3', 3, 'Fix login bug',        'Login fails on retry',       'proj-t1-a', 'test1', 'test1', 'test1', 1700000003000, 1700000003000, 0, 0, FALSE),
    ('iss-t1b-1', 'T1B-1', 1, 'Design landing page',  'Landing page mockups',        'proj-t1-b', 'test1', 'test1', 'test1', 1700000011000, 1700000011000, 0, 0, FALSE),
    ('iss-t1b-2', 'T1B-2', 2, 'Implement dark mode',  'Dark theme support',          'proj-t1-b', 'test1', 'test1', 'test1', 1700000012000, 1700000012000, 0, 0, FALSE),
    ('iss-t1b-3', 'T1B-3', 3, 'API documentation',    'Document public API',         'proj-t1-b', 'test1', 'test1', 'test1', 1700000013000, 1700000013000, 0, 0, FALSE),
    ('iss-t2a-1', 'T2A-1', 1, 'Database migration',   'Move to new schema',          'proj-t2-a', 'test2', 'test2', 'test2', 1700000021000, 1700000021000, 0, 0, FALSE),
    ('iss-t2a-2', 'T2A-2', 2, 'Add caching layer',    'Redis cache for reads',       'proj-t2-a', 'test2', 'test2', 'test2', 1700000022000, 1700000022000, 0, 0, FALSE),
    ('iss-t2a-3', 'T2A-3', 3, 'Performance tuning',   'Reduce API latency',          'proj-t2-a', 'test2', 'test2', 'test2', 1700000023000, 1700000023000, 0, 0, FALSE),
    ('iss-t2b-1', 'T2B-1', 1, 'User dashboard',       'Dashboard widgets',           'proj-t2-b', 'test2', 'test2', 'test2', 1700000031000, 1700000031000, 0, 0, FALSE),
    ('iss-t2b-2', 'T2B-2', 2, 'Export reports',       'CSV/PDF export',              'proj-t2-b', 'test2', 'test2', 'test2', 1700000032000, 1700000032000, 0, 0, FALSE),
    ('iss-t2b-3', 'T2B-3', 3, 'Email notifications',  'Notify on issue updates',     'proj-t2-b', 'test2', 'test2', 'test2', 1700000033000, 1700000033000, 0, 0, FALSE)
ON CONFLICT (id) DO NOTHING;

-- 11. الأدوار: دور عضو المشروع المستخدم في نطاق كل مشروع
INSERT INTO roles (id, name, description, is_updatable, immutable) VALUES
    ('project_member', 'project_member', 'Member of a specific project (scoped role)', TRUE, FALSE)
ON CONFLICT (id) DO NOTHING;

-- 12. الصلاحيات: تُبذر الناقصة فقط ولا تُمسّ الصلاحيات الموجودة مسبقاً
INSERT INTO permissions (id, name, description, permission_entity_type, localized_permission_entity_type, operation, is_global)
VALUES
    ('project.read',  'project.read',  'Read projects',        'Project', 'Project', 'read',  FALSE),
    ('project.write', 'project.write', 'Manage projects',      'Project', 'Project', 'write', FALSE),
    ('issue.read',    'issue.read',    'Read issues',          'Issue',   'Issue',   'read',  FALSE),
    ('issue.write',   'issue.write',   'Create and update issues', 'Issue', 'Issue',  'write', FALSE),
    ('profile.read',  'profile.read',  'Read user profiles',   'User',    'User',    'read',  FALSE),
    ('search.read',   'search.read',   'Use search',           'Issue',   'Issue',   'read',  FALSE),
    ('self.write',    'self.write',    'Write self data',      'User',    'User',    'write', FALSE)
ON CONFLICT (id) DO NOTHING;

-- 13. ربط الدور بالصلاحيات (دور عضو المشروع: قراءة المشاريع والقضايا + العمل على القضايا)
INSERT INTO role_permissions (role_id, permission_id, name, description,
                              permission_entity_type, localized_permission_entity_type, operation, is_global)
SELECT 'project_member', p.id, p.name, p.description,
       p.permission_entity_type, p.localized_permission_entity_type, p.operation, p.is_global
FROM permissions p
WHERE p.id IN ('project.read', 'issue.read', 'issue.write')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- 14. تعيين الأدوار بنطاق المشروع: هذا هو مصدر "ملكية" المشروع لكل مستخدم.
--     test3 لا يُعطى أي دور، لذلك يجب أن يرى قائمة فارغة.
INSERT INTO assigned_roles (id, role_id, holder_type, holder_id, holder_name, scope_type, scope_id, scope_project_id)
VALUES
    ('ar-t1a', 'project_member', 'user', 'test1', 'test1', 'project', 'proj-t1-a', 'proj-t1-a'),
    ('ar-t1b', 'project_member', 'user', 'test1', 'test1', 'project', 'proj-t1-b', 'proj-t1-b'),
    ('ar-t2a', 'project_member', 'user', 'test2', 'test2', 'project', 'proj-t2-a', 'proj-t2-a'),
    ('ar-t2b', 'project_member', 'user', 'test2', 'test2', 'project', 'proj-t2-b', 'proj-t2-b')
ON CONFLICT (id) DO NOTHING;

-- 15. ذاكرة الصلاحيات المخبأة لكل مستخدم (is_global = FALSE لأن الدور مقيّد بنطاق مشروع)
INSERT INTO cached_permissions (id, user_id, is_global)
SELECT rp.permission_id, ar.holder_id, FALSE
FROM assigned_roles ar
JOIN role_permissions rp ON rp.role_id = ar.role_id
WHERE ar.holder_type = 'user' AND ar.scope_project_id IS NOT NULL
ON CONFLICT (id, user_id) DO NOTHING;

-- 16. ربط كل صلاحية مخبأة بالمشاريع التي مُنحت فيها للمستخدم
INSERT INTO cached_permission_projects (permission_id, project_id)
SELECT DISTINCT rp.permission_id, ar.scope_project_id
FROM assigned_roles ar
JOIN role_permissions rp ON rp.role_id = ar.role_id
WHERE ar.holder_type = 'user' AND ar.scope_project_id IS NOT NULL
ON CONFLICT (permission_id, project_id) DO NOTHING;
