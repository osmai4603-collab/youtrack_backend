# كائن `Permission` — شرح مفصل

هذا المستند يشرح بنية (Structure) كائن الصلاحية `Permission` وخواصّه
كما هي معرّفة في `internal/services/permissions/permissions.go`.

## نظرة عامة

الكائن `Permission` هو الوحدة الأساسية لتمثيل صلاحية واحدة في نظام RBAC.
كل صلاحية تُجسّد قدرةً واحدة (مثل قراءة مشروع أو إنشاء قضية/Issue أو حذف
تعليق) مع معلومات تعريفية عنها. جميع الصلاحيات مخزّنة في مصفوفة
`All` داخل نفس الملف، وتُستخدم في `permission_service.go` للتحقق من
امتلاك المستخدم لأيّ صلاحية.

## تعريف الكائن

```go
type Permission struct {
 ID                            string
 Name                          string
 Description                   string
 PermissionEntityType          Entity
 LocalizedPermissionEntityType Module
 IsGlobal                      bool
 Operation                     Operation
 Dependents                    []string
 Implies                       []string
}
```

---

## شرح الخصائص

### 1) `ID string`

المعرّف النصّي الفريد للصلاحية بصيغة `kebab-case` باللغة الإنجليزية،
مثال: `"read-project-basic"`.

- يُستخدم هذا المعرّف كقيمة ثابتة في النظام وفي جداول قاعدة البيانات
  (مثل جدول `permission_dependencies` الخاص بعلاقات الصلاحيات).
- لا يحتوي مسافات ولا أحرف كبيرة، ويمثّل «الاسم التقني» الثابت
  المستقر الذي تعتمد عليه التطبيقات والقواعد البرمجية.

### 2) `Name string`

الاسم الكامل للصلاحية بشكل يتّبع صيغة `Title Case` الإنجليزي
للعرض، مثال: `"Read Project Basic"`.

- غالبًا يكون مشتقًّا من `ID` عبر تحويله إلى صيغة قابلة للعرض، لكنه
  مفهوم مستقل يمكن ترجمته (Localization).
- **مهم في النظام الحالي:** خدمة الصلاحيات في `permission_service.go`
  تعتمد على `Name` (وليس `ID`) عند التحقق:
  - `HasPermission` تبحث عن `permission.Name` ضمن أسماء صلاحيات الدور.
  - `GetPermissionByName` تأخذ `Name` وترجع مؤشر `*Permission`.

### 3) `Description string`

وصف نصّي موجز باللغة الإنجليزية لوظيفة الصلاحية، مثال:
`"View basic project information and settings"`.

- صفة تعريفيّة بحتة تستخدم للعرض والتوثيق، ولا تؤثر على منطق التحقق.

### 4) `PermissionEntityType Entity`

نوع الكيان (Entity) الذي تعمل عليه الصلاحية، وهو قيمة من النوع
المعرّف `Entity` (سلسلة نصية). الكيانات المتاحة:

```go
type Entity string

const (
 ProjectEntity       Entity = "project"
 OrganizationEntity  Entity = "organization"
 UserEntity          Entity = "user"
 SystemEntity        Entity = "system"
 IssueEntity         Entity = "issue"
 AttachmentEntity    Entity = "attachment"
 CommentEntity       Entity = "comment"
 VisibilityEntity    Entity = "visibility"
 IssueWorkItemEntity Entity = "issue-work-item"
 ArticleEntity       Entity = "article"
 AppEntity           Entity = "app"
)
```

مثال: صلاحية `PermIssueReadIssue` لها `PermissionEntityType: IssueEntity`
أي أنها تخصّ كيان القضايا.

### 5) `LocalizedPermissionEntityType Module`

تمثيل «مترجم/قابل للعرض» لنوع الكيان، ويمثّل اسم الوحدة (Module)
التي تنتمي إليها الصلاحية. أنواع الوحدات:

```go
type Module string

const (
 ProjectModule       Module = "Project"
 OrganizationModule  Module = "Organization"
 UserModule          Module = "User"
 SystemModule        Module = "System"
 IssueModule         Module = "Issue"
 AttachmentModule    Module = "Attachment"
 CommentModule       Module = "Comment"
 VisibilityModule    Module = "Visibility"
 IssueWorkItemModule Module = "Issue Work Item"
 ArticleModule       Module = "Article"
 PppModule           Module = "App"
)
```

مثال: صلاحية `PermVisibilityOverrideVisibilityRestrictions` لها
`PermissionEntityType: IssueEntity` (تعمل على كيان القضية) بينما
`LocalizedPermissionEntityType: VisibilityModule` (تُعرض ضمن وحدة
الرؤية Visibility).

### 6) `IsGlobal bool`

تحديد ما إذا كانت الصلاحية **عامة** (Global) أم **مرتبطة بمشروع/سياق**:

- `true`: الصلاحية عامة ولا تُقيَّد بمشروع محدد — أبرز مثال عليها
  صلاحيات المستخدم/النظام/المؤسسة/التطبيق مثل `PermUserReadUserBasic`
  و`PermSystemLowLevelAdminRead` و`PermOrganizationReadOrganization`
  و`PermAppReadAppContent`.
- `false`: الصلاحية مرتبطة بمشروع أو مورد ضمن نطاقه (مثل قراءة
  قضية، تعديل تعليق، إضافة مرفق...)، ومعظم صلاحيات Issue/Attachment/
  Comment/Article/WorkItem تأخذ هذه القيمة.

> في التهيئة بقاعدة البيانات (`migrations/000016_rbac_seed.up.sql`)
> تُترجم هذه القيمة إلى العمود `is_global`.

### 7) `Operation Operation`

العملية (العمل الإجرائي) التي تسمح بها الصلاحية، وهي قيمة من النوع
المعرّف `Operation`:

```go
type Operation string

const (
 ReadOperation     Operation = "read"
 CreateOperation   Operation = "create"
 UpdateOperation   Operation = "update"
 DeleteOperation   Operation = "delete"
 LinkOperation     Operation = "link"
 OverrideOperation Operation = "override"
 ApplyOperation    Operation = "apply"
)
```

| العملية | القيمة | المعنى |
| --------- | -------- | -------- |
| `ReadOperation` | `read` | قراءة/مشاهدة الموارد |
| `CreateOperation` | `create` | إنشاء موارد جديدة |
| `UpdateOperation` | `update` | تعديل/تحديث الموارد |
| `DeleteOperation` | `delete` | حذف الموارد |
| `LinkOperation` | `link` | إنشاء/إدارة روابط (مثل ربط القضايا مع بعضها) |
| `OverrideOperation` | `override` | تجاوز القيود (مثل قيود الرؤية) |
| `ApplyOperation` | `apply` | تطبيق أوامر (مثل أوامر القضايا بصمت) |

### 8) `Dependents []string` — المتفرّعون (من يعتمد عليك)

قائمة معرّفات الصلاحيات التي **تعتمد على هذه الصلاحية** كمقدّمة/شرط
أساسي (`Prerequisite`). بصيغة أوضح: **إذا كانت الصلاحية B موجودة في
`dependents` للصلاحية A، فهذا يعني أن منح B يستلزم منح A أولًا** —
أي أن A هي «الأساس» الذي بُنيت عليه B.

الاتجاه بالمصطلحات:

> `A.dependents = [B, C, ...]`  ⇒  **A هي المقدمة، و B و C فرعان يعتمدان عليها**.

#### مثال محلّل: `read-project-basic`

```go
PermProjectReadProjectBasic = Permission{
    Dependents: []string{  // 18 صلاحية تعتمد عليها
        "read-project-full",
        "read-issue",
        "read-issue-private-fields",
        "create-issue",
        "view-watchers",
        "view-voters",
        "read-article",
        "create-article",
        "update-article",
        "delete-article",
        "read-article-comment",
        "create-article-comment",
        "update-article-comment",
        "delete-article-comment",
        "update-issue-private-fields",
        "override-visibility-restrictions",
        "update-project",
        "delete-project",
    },
}
```

المعنى: أي صلاحية من هذه الـ18 تتطلب أن يكون المستخدم يملك
`read-project-basic` ضمنًا، لأنها كلها تعمل داخل نطاق مشروع ولا
يمكن استخدامها دون القدرة على «رؤية أساسيات المشروع».

#### قائمة كاملة بالصلاحيات التي تملك `Dependents` (غير فارغة)

| الصلاحية (A) | متفرّعوها `dependents` |
| --------------- | ------------------------- |
| `read-project-basic` | 18 صلاحية (انظر أعلاه) |
| `read-project-full` | `update-project`, `delete-project` |
| `read-organization` | `update-organization`, `delete-organization` |
| `update-self` | `update-user` |
| `read-user-basic` | `read-user-details`, `update-user`, `delete-user` |
| `read-user-details` | `update-user`, `delete-user` |
| `low-level-admin-read` | `low-level-admin-write` |
| `read-issue-private-fields` | `update-issue-private-fields`, `override-visibility-restrictions` |
| `update-issue` | `update-issue-private-fields` |
| `read-issue-comment` | `update-not-own-issue-comment`, `delete-not-own-comment-and-permanent-comment-delete` |
| `read-article-comment` | `create-article-comment`, `update-article-comment`, `delete-article-comment` |
| `read-article` | `create-article`, `update-article`, `delete-article`, `read-article-comment`, `create-article-comment`, `update-article-comment`, `delete-article-comment` |
| `read-work-item` | `update-not-own-work-item` |
| `update-work-item` | `update-not-own-work-item` |
| `create-work-item` | `create-not-own-work-item` |
| `read-app-content` | `update-app-content` |

---

### 9) `Implies []string` — المستلزمات (من تمنحه)

قائمة معرّفات الصلاحيات التي **تُمنح تلقائيًا عند منح هذه الصلاحية**
(أي أنها «تتضمّن» صلاحيات أخرى). الاتجاه بالمصطلحات:

> `A.implies = [X, Y, ...]`  ⇒  **منح A يمنح حاملها X و Y أيضًا**.

بمعنى آخر: هي عكسيّة `dependents` تمامًا — النقطة هنا أن `A` هي
«الأقوى» أو «الأوسع»، و `X`/`Y` هي صلاحيات «أضعف/أضيق» تُمنح ضمنًا.

#### مثال محلّل: `update-user`

```go
PermUserUpdateUser = Permission{
    Implies: []string{
        "read-user-details",  // أوسع، تغطيها
        "update-self",        // مستخدم هذا الدور يفعلها على نفسه
        "read-user-basic",    // الأساس نفسه
    },
}
```

المعنى: من يملك صلاحية `update-user` (تعديل حسابات المستخدمين) يُعتبر
ضمنًا قادرًا على قراءة تفاصيل أي مستخدم، وتعديل حسابه الشخصي، وقراءة
المعلومات الأساسية للمستخدمين.

#### قائمة كاملة بالصلاحيات التي تملك `Implies` (غير فارغة)

| الصلاحية (A) | المستلزمات `implies` |
| --------------- | ----------------------- |
| `read-project-full` | `read-project-basic` |
| `update-project` | `read-project-full`, `read-project-basic` |
| `delete-project` | `read-project-full`, `read-project-basic` |
| `update-organization` | `read-organization` |
| `delete-organization` | `read-organization` |
| `read-user-details` | `read-user-basic` |
| `update-user` | `read-user-details`, `update-self`, `read-user-basic` |
| `delete-user` | `read-user-details`, `read-user-basic` |
| `low-level-admin-write` | `low-level-admin-read` |
| `read-issue` | `read-project-basic` |
| `read-issue-private-fields` | `read-project-basic` |
| `create-issue` | `read-project-basic` |
| `update-issue-private-fields` | `read-issue-private-fields`, `update-issue`, `read-project-basic` |
| `view-watchers` | `read-project-basic` |
| `view-voters` | `read-project-basic` |
| `update-not-own-issue-comment` | `read-issue-comment` |
| `delete-not-own-comment-and-permanent-comment-delete` | `read-issue-comment` |
| `read-article` | `read-project-basic` |
| `create-article` | `read-article`, `read-project-basic` |
| `update-article` | `read-article`, `read-project-basic` |
| `delete-article` | `read-article`, `read-project-basic` |
| `read-article-comment` | `read-article`, `read-project-basic` |
| `create-article-comment` | `read-article-comment`, `read-article`, `read-project-basic` |
| `update-article-comment` | `read-article-comment`, `read-article`, `read-project-basic` |
| `delete-article-comment` | `read-article-comment`, `read-article`, `read-project-basic` |
| `override-visibility-restrictions` | `read-issue-private-fields`, `read-project-basic` |
| `update-not-own-work-item` | `read-work-item`, `update-work-item` |
| `create-not-own-work-item` | `create-work-item` |
| `update-app-content` | `read-app-content` |

---

### العلاقة العكسية بين `Dependents` و `Implies`

الخاصيتان تصفان **نفس العلاقة من طرفين متعاكسين**؛ لأي زوج
(صلاحية أقوى ⇔ صلاحية أساسية) تكون القاعدة:

> إذا احتوت `X.implies` على `Y`  ⇔  يوجد `X` في `Y.dependents`.

أمثلة حقيقية من الملف تؤكد هذه المرآتية:

| `X.implies` (يحتوي Y) | `Y.dependents` (يحتوي X) |
| ----------------------- | --------------------------- |
| `update-project` → `read-project-full` | `read-project-full` → `update-project` |
| `update-project` → `read-project-basic` | `read-project-basic` → `update-project` |
| `update-organization` → `read-organization` | `read-organization` → `update-organization` |
| `read-user-details` → `read-user-basic` | `read-user-basic` → `read-user-details` |
| `update-user` → `read-user-details` | `read-user-details` → `update-user` |
| `update-user` → `read-user-basic` | `read-user-basic` → `update-user` |
| `low-level-admin-write` → `low-level-admin-read` | `low-level-admin-read` → `low-level-admin-write` |
| `update-issue-private-fields` → `read-issue-private-fields` | `read-issue-private-fields` → `update-issue-private-fields` |
| `create-article-comment` → `read-article-comment` | `read-article-comment` → `create-article-comment` |
| `update-app-content` → `read-app-content` | `read-app-content` → `update-app-content` |

---

### التتابع (السلاسل) بين الصلاحيات `Transitive Chains`

بما أن الصلاحيات قد تعبّر عن «أقوى» منها أو «أساس» لها، يمكن أن تنشأ
**سلسلة تتابع** غير مباشرة:

```text
update-project  ⇐implies⇒  read-project-full  ⇐implies⇒  read-project-basic
```

أي أن منح `update-project` ينقل تلقائيًا:
`update-project` ⇒ `read-project-full` ⇒ `read-project-basic`
(ثلاث صلاحيات كاملة من صلاحية واحدة). نفس النمط نجده في:
`update-user` ⇒ `read-user-details` ⇒ `read-user-basic`.

هذه التتابعات هي ما يجعل التحقق عبر الحلقات (`Dependents`/`Implies`)
ضروريًا في `HasPermission` بدل الاكتفاء بفحص الاسم المباشر.

---

### كيف تستخدم خدمة الصلاحيات الخاصيتين؟

الفحص في `HasPermission` (`permission_service.go:21`) يستخدم الخاصيتين
معًا بنفس الحلقة، وفي **اتجاهين** بالنسبة للصلاحية المطلوبة:

```go
for _, perm := range permissions {
    if slices.Contains(permission.Dependents, perm.Name) ||
       slices.Contains(permission.Implies, perm.Name) {
        return true, nil
    }
}
```

خطوات التحقق كاملة:

1. جلب أسماء صلاحيات الأدوار (`roleIDs`) عبر `GetPermissionsByRoleIDs`.
2. إن أُرسل اسم الصلاحية المطلوبة (`permission.Name`) مباشرة في قائمة
   أسماء الصلاحيات ⇒ يمتلكها المستخدم، ولا يُرجع خطأً.
3. إن لم يكن الاسم المباشر موجودًا، تمرّ على كل صلاحية من صلاحيات
   المستخدم (`perm`) وتفحص اثنين من أسماءها:
   - **`permission.Dependents`**: هل صلاحية المستخدم يتفرّع عن الصلاحية
     المطلوبة؟ (أي أن الصلاحية المطلوبة «أساس» لصلاحية المستخدم —
     فامتلاكه الفرع يعني امتلاكه الأساس).
   - **`permission.Implies`**: هل صلاحية المستخدم من 'المستلزمات'
     التي تمنحها الصلاحية المطلوبة؟ (أي أن المستخدم يملك شيئًا نحتاج
     الصلاحية المطلوبة للوصول إليه).

> **الفارق الجوهري:** `Dependents` تتحقق «لأعلى» (المستخدم يملك فرعًا
> فأساسُه مُحتسب)، بينما `Implies` تتحقق «لأسفل» (المستخدم يملك
> صلاحيةً نحن نحتاجها ضمن امتلاكه صلاحيةً أوسع).

هذا يعني أن `Dependents` و `Implies` معًا تتيحان **الاستدلال غير
المباشر الكامل**: المستخدم قد «يمتلك» صلاحيةً دون أن تُسند إليه
بشكل صريح، فقط بامتلاكه صلاحيةً مرتبطة بها في أيّ من الاتجاهين —
سواء كانت صلاحيته «أقوى» (سُضبط عبر `implies`) أو «أضعف» (سُضبط
عبر `dependents`).

---

## جدول الصلاحيات المحدّدة (`All`)

يوجد حاليًا **53 صلاحية** مخزّنة في الملف، موزّعة على الوحدات التالية:

| الوحدة | الصلاحيات |
| -------- | ----------- |
| Project | `read-project-basic`, `create-project`, `read-project-full`, `update-project`, `delete-project` |
| Organization | `read-organization`, `update-organization`, `create-organization`, `delete-organization` |
| User Profile | `update-self` |
| User | `read-user-basic`, `read-user-details`, `update-user`, `create-user`, `delete-user` |
| System | `low-level-admin-read`, `low-level-admin-write` |
| Issue | `read-issue`, `read-issue-private-fields`, `update-issue`, `create-issue`, `delete-issue`, `link-issues`, `update-issue-private-fields`, `apply-commands-silently`, `view-watchers`, `update-watchers`, `view-voters` |
| Attachment | `add-attachment`, `update-attachment`, `delete-attachment` |
| Comment | `create-issue-comment`, `read-issue-comment`, `update-issue-comment`, `delete-issue-comment`, `update-not-own-issue-comment`, `delete-not-own-comment-and-permanent-comment-delete`, `read-article-comment`, `create-article-comment`, `update-article-comment`, `delete-article-comment` |
| Visibility | `override-visibility-restrictions` |
| Issue Work Item | `read-work-item`, `update-work-item`, `update-not-own-work-item`, `create-work-item`, `create-not-own-work-item` |
| Article | `read-article`, `create-article`, `update-article`, `delete-article` |
| App | `read-app-content`, `update-app-content` |

---

## خلاصة

| الخاصية | النوع | الغرض |
| --------- | ------- | ------- |
| `ID` | `string` | معرّف تقني فريد بصيغة `kebab-case` |
| `Name` | `string` | اسم قابل للعرض، يُعتمد عليه في التحقق |
| `Description` | `string` | وصف وظيفي تعريفي |
| `PermissionEntityType` | `Entity` | الكيان الذي تعمل عليه الصلاحية |
| `LocalizedPermissionEntityType` | `Module` | اسم الوحدة القابل للعرض |
| `IsGlobal` | `bool` | هل الصلاحية عامة أم مرتبطة بمشروع |
| `Operation` | `Operation` | العملية المسموحة (read/create/update/...) |
| `Dependents` | `[]string` | صلاحيات تعتمد على هذه الصلاحية (Prerequisites) |
| `Implies` | `[]string` | صلاحيات تُمنح تلقائيًا عند منح هذه الصلاحية |
