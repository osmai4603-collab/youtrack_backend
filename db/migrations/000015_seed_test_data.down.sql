-- 000015_seed_test_data.down.sql
-- التراجع عن بيانات الاختبار التي أضافها 000015_seed_test_data.up.sql.
--
-- يُحذف بترتيب عكسي لترتيب الإدراج (تفادياً لمخالفات المفاتيح الأجنبية).
-- التنفيذ يجري داخل معاملة واحدة من مُشغّل الـ migrations.
--
-- قاعدة السلامة المعتمدة هنا: لا يُحذف أي صف إلا إذا كان **مطابقاً تماماً** للصفوف التي
-- يُدرجها 000015_seed_test_data.up.sql. السبب: الجداول المشتركة
-- (permissions / roles / role_permissions / user_types / project_types)
-- قد تكون أُنشئت مسبقاً بواسطة cmd/seeder أو ببيانات إنتاجية، وحذفها
-- كان سيمسح بيانات لا علاقة لها ببيانات الاختبار.

-- 16. ربط الصلاحيات المخبأة بالمشاريع
DELETE FROM cached_permission_projects
WHERE permission_id IN ('project.read', 'project.write', 'issue.read', 'issue.write',
                        'profile.read', 'search.read', 'self.write')
  AND project_id IN ('proj-t1-a', 'proj-t1-b', 'proj-t2-a', 'proj-t2-b');

-- 15. ذاكرة الصلاحيات المخبأة (مقيّدة بالمستخدمين وبالصلاحيات المبوذرة)
DELETE FROM cached_permissions
WHERE user_id IN ('test1', 'test2', 'test3')
  AND id IN ('project.read', 'project.write', 'issue.read', 'issue.write',
             'profile.read', 'search.read', 'self.write');

-- 14. الأدوار المعيّنة بنطاق المشروع: الصفوف الأربعة المبوذرة فقط،
--     حتى لا يُحذف دور أُسند لاحقاً إلى أحد مستخدمي الاختبار.
DELETE FROM assigned_roles
WHERE id IN ('ar-t1a', 'ar-t1b', 'ar-t2a', 'ar-t2b');

-- 13. ربط دور project_member بالصلاحيات: الأزواج الثلاثة المبوذرة فقط
--     (لا تُحذف ربطت سابقة لنفس الدور).
DELETE FROM role_permissions
WHERE role_id = 'project_member'
  AND permission_id IN ('project.read', 'issue.read', 'issue.write')
  AND operation IN ('read', 'write');

-- 12. الصلاحيات المبوذرة: تُحذف فقط عند مطابقة الوصف المُدرج في الملف الصاعد،
--     فتبقى أي صلاحية سابقة بنفس المعرّف وبوصف مختلف.
DELETE FROM permissions
WHERE (id = 'project.read'  AND description = 'Read projects')              OR
      (id = 'project.write' AND description = 'Manage projects')            OR
      (id = 'issue.read'    AND description = 'Read issues')                OR
      (id = 'issue.write'   AND description = 'Create and update issues')  OR
      (id = 'profile.read'  AND description = 'Read user profiles')         OR
      (id = 'search.read'   AND description = 'Use search')                 OR
      (id = 'self.write'    AND description = 'Write self data');

-- 11. دور project_member: يُحذف فقط إذا كان مطابقاً لما أدخله الملف الصاعد
DELETE FROM roles
WHERE id = 'project_member'
  AND name = 'project_member'
  AND description = 'Member of a specific project (scoped role)';

-- 10. القضايا
DELETE FROM issues WHERE id IN ('iss-t1a-1', 'iss-t1a-2', 'iss-t1a-3',
                                'iss-t1b-1', 'iss-t1b-2', 'iss-t1b-3',
                                'iss-t2a-1', 'iss-t2a-2', 'iss-t2a-3',
                                'iss-t2b-1', 'iss-t2b-2', 'iss-t2b-3');

-- 9. فك ارتباط مجموعة الرؤية الافتراضية قبل حذف المجموعات
UPDATE projects SET default_visibility_group_id = NULL
WHERE id IN ('proj-t1-a', 'proj-t1-b', 'proj-t2-a', 'proj-t2-b');

-- 8. أعضاء المجموعات ثم المجموعات
DELETE FROM user_group_members
WHERE group_id IN ('grp-t1-a', 'grp-t1-b', 'grp-t2-a', 'grp-t2-b');

DELETE FROM user_groups
WHERE id IN ('grp-t1-a', 'grp-t1-b', 'grp-t2-a', 'grp-t2-b');

-- 7. أعضاء الفرق
DELETE FROM project_team_members
WHERE team_id IN ('team-t1-a', 'team-t1-b', 'team-t2-a', 'team-t2-b');

-- 5. المشاريع (بعد فك team_id لتفادي كسر مرجع projects.team_id)
UPDATE projects SET team_id = NULL
WHERE id IN ('proj-t1-a', 'proj-t1-b', 'proj-t2-a', 'proj-t2-b');

DELETE FROM projects
WHERE id IN ('proj-t1-a', 'proj-t1-b', 'proj-t2-a', 'proj-t2-b');

-- 6. ربط الفرق بالمشاريع ثم حذف الفرق
UPDATE project_teams SET project_id = NULL
WHERE id IN ('team-t1-a', 'team-t1-b', 'team-t2-a', 'team-t2-b');

DELETE FROM project_teams
WHERE id IN ('team-t1-a', 'team-t1-b', 'team-t2-a', 'team-t2-b');

-- 3. ملفات التعريف ثم المستخدمون
DELETE FROM user_profiles WHERE user_id IN ('test1', 'test2', 'test3');
DELETE FROM users WHERE id IN ('test1', 'test2', 'test3');

-- 1. أنواع المستخدمين والمشاريع
--    تُترك كما هي عمداً: هي صفوف مرجعية مشتركة (lookup) ينشئها أيضاً
--    cmd/seeder، ووجودها لا يمثل بيانات اختبار. مفاتيحها الأجنبية معرّفة
--    بـ ON DELETE SET NULL، فتركها لا يمنع إعادة تطبيق أي migration لاحق.
