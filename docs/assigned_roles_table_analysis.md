# تحليل جدول `assigned_roles`: ما هو ولماذا نحتاجه؟

> وثيقة تحليلية تشرح جدول الأدوار المُسندة `assigned_roles` في واضعة البيانات، ودوره في نظام
> التحكم بالوصول القائم على الأدوار (RBAC)، ولماذا كان وجوده ضرورياً لفصل "تعريف الدور"
> عن "إسناد الدور إلى شخص معيّن داخل نطاق معيّن".

---

## 1. نبذة مختصرة

جدول `assigned_roles` هو الجدول المحوري في نظام الصلاحيات، وهو **الوسيط (Join/Assignment Table)**
الذي يربط بين **الأدوار (`roles`)** وبين **حامليها (المستخدمين `users` أو المجموعات `user_groups`)**
مع تحديد **نطاق الإسناد (Scope)** (عام / مؤسسة / مشروع).

يُعرَّف في `migrations/000001_init_schema.up.sql` (سطر 581) كما يلي:

```sql
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
```

مقابل له في الكود: البنية `RoleAssignment` في
`internal/storage/permissions/permission_repository.go:14`.

---

## 2. موقعه في معمارية RBAC

يعتمد النظام مبدأً أساسياً ذكره التوثيق السابق (`docs/youtrack_permissions_and_global_scope.md`):

> الصلاحيات **لا تُمنح للمستخدم مباشرة مطلقاً**، بل تُسند دائماً عبر أدوار مُسندة في نطاق محدد.

سلسلة العلاقة هي:

```
[المستخدم / المجموعة (Holder)]
        |
        |  (سجل إسناد في assigned_roles: من؟ مَن؟ أين؟)
        v
[Rolle: roles]  -------------->  [صلاحيات: role_permissions]  -------------->  [Permission: permissions]
```

- `roles`: يصف **الأدوار المجرّدة** (contributor / project-admin / system-admin) وما تتضمنه كل واحدة من صلاحيات.
- `role_permissions`: يربط كل دور بصلاحياته (دور → صلاحيات).
- `assigned_roles`: يربط كل دور بـ**حامل ونطاق** (مَن يحمل الدور، وفي أي نطاق).

دور الجدول هنا هو **مصفوفة الإسناد (Assignment Matrix)** التي تجعل الصلاحية نافذة فعلياً لشخص معيّن، لا مجرّد تعريف نظري في الكتالوج.

---

## 3. شرح الأعمدة عموداً عموداً

| العمود | النوع | المعنى الوظيفي |
| :--- | :--- | :--- |
| `id` | `VARCHAR(20) PK` | معرّف فريد لسجل الإسناد نفسه (يُستخدم لعملية الإلغاء `RevokeRole`). |
| `role_id` | `VARCHAR(100) FK → roles(id)` | أيُّ دور يُسند (المصدر). الحذف من `roles` يحذف الإسنادات تلقائياً (`ON DELETE CASCADE`). |
| `audit_target_id` | `VARCHAR(100)` | هدف التدقيق/التتبّع (مرآة لمفهوم `audit` في YouTrack) — يُستخدم لتسجيل مَن فعل ماذا. |
| `holder_type` | `VARCHAR(50)` | نوع الحامل: `user` أو `group`. يحدد كيف يُفسَّر `holder_id`. |
| `holder_id` | `VARCHAR(20)` | معرّف الحامل: `users.id` إذا كان المستخدم، أو `user_groups.id` إذا كانت مجموعة. |
| `holder_name` | `VARCHAR(255)` | تسجيل لحظي لاسم الحامل وقت الإسناد (للعرض/التدقيق دون الحاجة إلى JOIN). |
| `scope_type` | `VARCHAR(50)` | نطاق الإسناد: `global`, `organization`, `project` (يتطابق مع `ScopeType` في `permission_service.go:11`). |
| `scope_id` | `VARCHAR(20)` | معرّف نطاق عام/مرن (أي معرّف السياق في حالات خاصة). |
| `scope_project_id` | `VARCHAR(20) FK → projects(id)` | إذا كان النطاق `project`: أيُّ مشروع. |
| `scope_organization_id` | `VARCHAR(20) FK → organizations(id)` | إذا كان النطاق `organization`: أيُّ مؤسسة. |

**ملاحظة معمارية:** العمودان `scope_id` و `scope_type` إلى جانب `scope_project_id` و
`scope_organization_id` أوجدها النظام لدعم النطاقات الثلاثة (Global / Organization / Project)
بشكل مرن، مع السماح بأن تُفسَّر `scope_id` بشكل سياقي اعتماداً على `scope_type`
(كما يظهر في استعلامات `permission_repository.go` في المقارنات المزدوجة `scope_id = $n OR scope_project_id = $n`).

---

## 4. أنواع النطاقات (Scopes) الثلاثة

النظام يعرّف ثلاثة نطاقات في `internal/services/permissions/permission_service.go:11`:

```go
ScopeGlobal       ScopeType = "global"
ScopeOrganization ScopeType = "organization"
ScopeProject      ScopeType = "project"
```

| النطاق | المعنى | أمثلة على أدوار تُسند فيه |
| :--- | :--- | :--- |
| `global` | يسري على النظام ككل | `system-admin` (وصلاحيات User/System/Org العامة) |
| `organization` | يسري على كافّة مشاريع المؤسسة | `project-admin` لمؤسسة معيّنة |
| `project` | يسري على مشروع واحد محدد | `contributor` داخل مشروع Backend |

**لماذا هذا مهم؟** نفس الدور قد يُسند لمستخدم واحد في **عدة نطاقات** وبتأثيرات مختلفة.
لذلك كان لا بد من كيان (جدول) مستقل يخزّن "الدور + الحامل + النطاق" كسجلّ واحد مستقل،
بدلاً من حقل إسناد واحد داخل جدول `users`.

---

## 5. كيف يُستخدم الجدول في الكود؟

### أ) استرجاع صلاحيات المستخدم (الأهم)

الدالة `GetPermissionsForUser` في `permission_repository.go:64`:

1. تجمع أدوار المستخدم من `assigned_roles` (إما مباشرةً عبر `holder_type = 'user' and holder_id = $1`
   أو غير مباشرة عبر مجموعاته `group_members` عندما يكون `holder_type = 'group'`).
2. ترشّحها حسب النطاق المطلوب (`global` أو `organization` أو `project`).
3. JOIN مع `role_permissions` ثم `permissions` لاستخراج الصلاحيات الفعلية.
4. من ثم يقرّر `PermissionService.IsAllowed` في `permission_service.go:47` السماح أو الرفض.

### ب) استرجاع الأدوار فقط

`GetRolesByUserID` في `permission_repository.go:200` — نفس منطق الترشيح بالنطاق
لكنه يعيد قائمة `role_id` فقط (تُستخدم للتحققات المباشرة القديمة عبر `HasPermission`).

### ج) الإسناد والإلغاء

- `AssignRole` في `permission_repository.go:257` — يُدرج سجلاً جديداً في `assigned_roles`
  (يُحقّق أن `id`, `role_id`, `holder_id` غير فارغة، ويفترض `global`/`user` افتراضياً).
- `RevokeRole` في `permission_repository.go:306` — `DELETE FROM assigned_roles WHERE id = $1`.

### د) عرض الإسنادات

`GetAssignedRoles` في `permission_repository.go:327` — يسترجع كافة إسنادات حامل معيّن
(تُستخدم لإدارة الأدوار من لوحة التحكم).

---

## 6. لماذا نحتاج هذا الجدول؟ (المبرّرات)

1. **فصل التعريف عن الإسناد:** الأدوار تُعرَّف مرة واحدة في `roles`، لكن أن يُكرّر دورُ ما
   لـ المُستخدمين بأعداد كبيرة وفي نطاقات مختلفة يستلزم جدول علاقة وسيط. لا يمكن وضع حقل
   `role_id` داخل `users` لأن المستخدم قد يحمل أدواراً متعددة ونطاقات مختلفة في آنٍ واحد.

2. **دعم النطاقات الثلاثة:** بدون هذا الجدول لا يوجد مكان يخزّن "هذا المستخدم هو
   contributor **في مشروع X فقط**، وليس في مشروع Y". الأعمدة `scope_*` تمنح هذه القدرة
   التعبيرية في سجل واحد قابل للتكرار.

3. **دعم الإسناد للمجموعات (`group`):** إسناد دور لمجموعة يمنحه تلقائياً لأعضائها.
   جدول الإسناد يجيز `holder_type = 'group'`، والنظام النهائي (الاستعلام في سطر 82-84)
   يطبّق هذا الانتقال عبر `group_members`. ارتداء ذلك في جدول `users` وحده كان سيحرمنا من هذه الطبقة.

4. **السيادة على الحذف والتدقيق:** `ON DELETE CASCADE` على `role_id`, `scope_project_id`,
   `scope_organization_id` يضمن أنه عند حذف دور أو مشروع أو مؤسسة تُحذف كل إسناداتها
   تلقائياً، فلا تتبقّى إسنادات يتيمة. كما أن `id` المستقل يعطي مرجعاً نظيفاً
   لعملية إلغاء محدّدة (`RevokeRole`).

5. **تسجيل باسم وهدف التدقيق:** تخزين `holder_name` و `audit_target_id` يجعل سجلّ الإسناد
   قابلاً للعرض والتدقيق دون ربط إضافي (denormalization مقصودة لأغراض الأداء/القراءة).

6. **تجنّب تخزين الصلاحيات مباشرة:** المنهجية في هذا المشروع (كما في YouTrack)
   تمنع إسناد صلاحيات أولية إلى مستخدم مباشرة؛ وحده `assigned_roles` يسمح بالوساطة عبر الأدوار
   وهو جهة الحقيقة الوحيدة الواردة في استعلامات التحقق.

---

## 7. العلاقات مع الجداول المجاورة

```text
roles(id) ─────┐  ON DELETE CASCADE
               v
        assigned_roles ── role_id ──► roles
        ── scope_project_id ──► projects
        ── scope_organization_id ──► organizations
        ── holder_id + holder_type (دلالي، بلا FK صريح) ──► users / user_groups

group_members(user_id, group_id)  ──► تُستخدم عند التوسعة من مجموعة إلى أعضائها
```

> **قرار تصميمي:** `holder_id` بلا `FOREIGN KEY` صريح لأن الحامل قد يكون مستخدماً أو
> مجموعة؛ فالربط يتم برمجياً عبر `holder_type` (سطر 81-84). هذا نمط شائع في جداول
> الإسناد الثنائي Polymorphic Assignment، ويُقابله `role_permissions → permissions`
> للصلاحيات.

---

## 8. مقارنة قصيرة مع البدائل المفترضة

| البديل | لماذا لم يُعتمد؟ |
| :--- | :--- |
| **حقل `role_id` داخل `users`** | مستخدم واحد = أدوار متعددة ونطاقات متعددة؛ لا يصلح لعلاقات One-to-Many محورية. |
| **جدول `role_permissions` فقط لا `assigned_roles`** | `role_permissions` توضح صلاحيات الدور، لكنها لا تقول *مَن* يحمل الدور *وفي أي نطاق*. بدون `assigned_roles` لا توجد "جهة إسناد". |
| **تخزين الصلاحيات مباشرة في `users`** | يخالف مبدأ RBAC الأساسي ويُصعّب إلغاء الأدوار وتحديثها مركزياً، ويفقد طبقة النطاقات. |

---

## 9. الخلاصة

`assigned_roles` هو **جهة الحقيقة (Source of Truth) لعملية الإسناد** في النظام:

- يربط الأدوار (المُعرَّفة في `roles`) بالحاملين (مستخدم/مجموعة) والنطاق (عام/مؤسسة/مشروع).
- هو ما يجعل صلاحية معيّنة مُتحققاً منها فعلياً لشخص معيّن في سياق معيّن،
  عبر الاستعلامات في `internal/storage/permissions/permission_repository.go`.
- وجوده يمنحنا: أدواراً متعددة لكل مستخدم، إسناداً جماعياً عبر المجموعات،
  نطاقات ثلاثية المستويات، وتنظيفاً تلقائياً عند حذف الأدوار/المشاريع/المؤسسات.

**باختصار:** بدون `assigned_roles` ستملك "كتالوج صلاحيات وأدوار" جميلاً لكنه معزول
عن المستخدمين وعن الفضاءات الواقعية (المنظمات والمشاريع) التي تُطبَّق فيها تلك الصلاحيات.
