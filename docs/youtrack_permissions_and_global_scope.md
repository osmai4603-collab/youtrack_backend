# توثيق شامل: نظام صلاحيات YouTrack، الخاصية Global، الكينونات، ونطاقات الأمان

هذا المستند يقدم دراسة تحليليّة مفصّلة حول **نظام الصلاحيات والأدوار في منصّة YouTrack (من تطوير JetBrains)**، مستنداً إلى المصادر الرسمية لنظام YouTrack وHub، ومطابقاً إياها مع المعمارية البرمجية وقاعدة البيانات المعتمدة في مشروعنا هذا.

---

## 1. كيف يعمل نظام الصلاحيات في YouTrack؟ (Role-Based Access Control - RBAC)

يعتمد YouTrack نظام للتحكم بالوصول قائم على الأدوار (RBAC). يتميز هذا النظام بالتالي:

### أ) الهيكل الأساسي للوصول

1. **الصلاحية (Permission)**: هي التخويل الذري (Atomic Grant) الممنوح للقيام بعملية محددة (مثل: قراءة قضية، تعديل مستخدم، إنشاء مشروع). الصلاحيات **ثابتة ومعرّفة مسبقاً** في كود النظام وقاعدة البيانات، ولا يمكن للمستخدمين إنشاؤها أو تعديل معرفاتها.
2. **الدور (Role)**: هو حزمة أو مجموعة منطقية تجمع صلاحيات متعددة معاً (مثل دور: *Developer*, *Project Admin*, *System Admin*, *Reporter*).
3. **المستخدم / المجموعة (User / Group)**: يُسند الدور للمستخدم مباشرة أو للمجموعة التي ينتمي إليها.
4. **النطاق (Scope)**: عند إسناد دور لمستخدم أو مجموعة، **يجب تحديد نطاق الإسناد**. قد يكون النطاق عاماً على مستوى النظام ككل (Global Scope)، أو محدداً بمشروع معين (Project Scope) أو مؤسسة (Organization Scope).

> **مبدأ أساسي:** الصلاحيات **لا تُمنح للمستخدم مباشرة مطلقاً**، بل تُسند دائماً عبر **أدوار (Roles)** مُسندة في **نطاق (Scope)** محدد.

```text
[المستخدم / المجموعة]  --->  [تعيين الدور (Assigned Role)]  --->  [الدور (Role)]  --->  [الصلاحيات (Permissions)]
                                        |
                                        +---> [النطاق (Scope): Global / Project / Organization]
```

### ب) علاقات الاعتماد والاستلزام بين الصلاحيات (Dependencies & Implications)

الصلاحيات في YouTrack ليست معزولة، بل تترابط في شجرة اعتمادية:

- **`Implies` (المستلزمات)**: منح صلاحية قوية يمنح ضمنياً صلاحيات أقل قوة (مثال: منح `update-project` يمنح تلقائياً `read-project-full` و `read-project-basic`).
- **`Dependents` (المتفرّعون / الشروط المسبقة)**: استخدام صلاحية فرعية يستلزم وجود الصلاحية الأساسية كشرط مسبق (مثال: صلاحية `read-project-basic` تعتبر شرطاً أساسياً لـ 18 صلاحية أخرى داخل المشروع).

---

## 2. ما الذي تعنيه الخاصية `global` (`isGlobal = true`)؟

في نظام YouTrack، تُصنف كل صلاحية عبر خاصية بولينية: `isGlobal` (`IsGlobal`).

### أ) التعريف الوظيفي لـ `global`

الصلاحية التي تملك خاصية `isGlobal = true` هي **صلاحية عامة على مستوى النظام بكامله (System-Wide Permission)**.

- **الصلاحية العامة (`isGlobal = true`)**: هي صلاحية تتعلق بوظائف وموارد لا تتبع مشروعاً بعينه، بل تخص النظام ككل، أو إدارة الحسابات والمؤسسات والتهيئة الكلية.
- **الصلاحية غير العامة (`isGlobal = false`)**: هي صلاحية سياقية (Contextual / Project-Based Permission) ترتبط بمشروع محدد أو الموارد التابعة له (مثل القضايا، التعليقات، المقالات، تتبع الوقت داخل المشروع).

### ب) المقارنة بين الصلاحية العامة والمرتبطة بمشروع

| وجه المقارنة | الصلاحية العامة (`isGlobal = true`) | الصلاحية المرتبطة بمشروع (`isGlobal = false`) |
| :--- | :--- | :--- |
| **طبيعة المورد** | موجه للنظام الكلي، الحسابات، أو المؤسسة | موجه لمشروع محدد أو عناصر داخل المشروع |
| **النطاق المطلوب للإسناد** | تتطلب دوراً مُسنداً في **النطاق العام (Global Scope)** | تتطلب دوراً مُسنداً في **نطاق ذلك المشروع (Project Scope)** (أو دوراً عاماً يغطي كافة المشاريع) |
| **أمثلة حقيقية** | `create-project`, `create-user`, `low-level-admin-read`, `read-organization` | `read-issue`, `create-issue-comment`, `add-attachment`, `update-project` |
| **تأثير المشروع** | لا تتأثر بالمشروع الذي يتواجد فيه المستخدم | ترتبط القبول أو الرفض بالمشروع المعني بالطلب |

---

## 3. ما الذي تعنيه `global`؟ هل يعني ذلك "لا يحتاج إلى تحقق من صلاحية"؟

### الإجابة القاطعة: **قطعاً لا (No, Absolutely Not!)**

هناك فهم خاطئ شائع يظن أن كلمة "Global" تعني "متاحة للجميع بدون قيود" أو "تتخطى فحص الأمان". هذا غير صحيح إطلاقاً.

### أ) الفارق بين "نوع الفحص" و "نطاق الفحص"

1. **التحقق من الصلاحية (Permission Check)**: هو العملية الأمنية التي تتأكد هل يملك المستخدم الحق البرمجي للقيام بالعملية أم لا. **كل طلب يحتاج تحققاً أمنياً سواء كان عاماً أو مخصصاً**.
2. **نطاق الفحص (Permission Scope)**: يحدد أين نبحث عن إسناد الصلاحية:
   - **في الصلاحية العامة (`isGlobal = true`)**: يفحص النظام هل يملك المستخدم **دوراً عاماً (Global Role Assignment)** يتضمن هذه الصلاحية.
   - **في الصلاحية المخصصة (`isGlobal = false`)**: يفحص النظام هل يملك المستخدم دوراً في **مشروع محدد (Project-Specific Role Assignment)** يمنحه هذه الصلاحية لذلك المشروع بالتحديد.

### ب) ماذا يحدث عند تنفيذ عملية تطلب صلاحية عامة؟

عندما يحاول مستخدم تنفيذ عملية عامة (مثل `create-user` أو `create-organization`):

1. يتلقى الميدلوير أو الخدمة الطلب ويستخرج معرف المستخدم (`userID`).
2. يجلب النظام كافة الأدوار المُسندة للمستخدم في **النطاق العام (Global Scope)**.
3. يفحص النظام هل الصلاحية المطلوبة (أو صلاحية أعلى تسلتزمها عبر `Implies`/`Dependents`) موجودة ضمن الأدوار العامة للمستخدم.
4. **إذا لم يملك المستخدم الصلاحية في النطاق العام**، يتم رفض الطلب فوراً بـ **`403 Forbidden` (HTTP Status Code 403)**.

---

## 4. من هي الكينونات (Entities) التي تطبق عليها خاصية `global`؟

تتوزع الصلاحيات في YouTrack على كيانات (Entities) وحدود معمارية. الجدول التالي يوضح الكيانات المتاحة وموقف خاصية `isGlobal` منها:

### أ) الكينونات التي تطبق عليها خاصية `global` (`isGlobal = true`)

1. **الكينونة `system` (النظام)**:
   - كافة صلاحيات الإدارة المنخفضة المستوى والتهيئة العامة للنظام.
   - الأمثلة: `low-level-admin-read`, `low-level-admin-write`.
2. **الكينونة `user` (المستخدمون والحسابات)**:
   - إنشاء الحسابات، تعديل بيانات المستخدمين، حذف الحسابات، وعرض التفاصيل الشاملة على مستوى النظام.
   - الأمثلة: `create-user`, `update-user`, `delete-user`, `read-user-basic`, `read-user-details`, `update-self`.
3. **الكينونة `organization` (المؤسسات والهيكل التنظيمي)**:
   - إدارة المؤسسات والمجموعات الهيكلية التي تحتضن المشاريع.
   - الأمثلة: `create-organization`, `read-organization`, `update-organization`, `delete-organization`.
4. **الكينونة `app` (التطبيقات والملحقات العامة)**:
   - قراءة وتحديث محتوى التطبيقات والملحقات المضافة للنظام ككل.
   - الأمثلة: `read-app-content`, `update-app-content`.
5. **الكينونة `project` (في عملية الإنشاء حصراً)**:
   - صلاحية **`create-project`** هي الصلاحية الوحيدة في كيان المشروع التي تحمل `isGlobal = true`. والسبب المنقي: **إنشاء مشروع جديد يتم على مستوى النظام ككل قبل وجود المشروع نفسه**، فلا يمكن ربط إنشاء مشروع بنطاق مشروع غير موجود بعد!

---

### ب) الكينونات التي لا تطبق عليها خاصية `global` (`isGlobal = false` - مرتبطة بالمشاريع)

1. **الكينونة `issue` (القضايا/المهام)**:
   - جميع عمليات القضايا محصورة داخل مشروع معُين (`read-issue`, `create-issue`, `update-issue`, `delete-issue`, `link-issues`, `read-issue-private-fields`, `update-issue-private-fields`, `view-watchers`, `update-watchers`, `view-voters`, `apply-commands-silently`, `override-visibility-restrictions`).
2. **الكينونة `comment` (التعليقات)**:
   - التعليقات على القضايا أو المقالات داخل مشروع معين (`create-issue-comment`, `read-issue-comment`, `update-issue-comment`, `delete-issue-comment`, `update-not-own-issue-comment`, `delete-not-own-comment-and-permanent-comment-delete`, `read-article-comment`, `create-article-comment`, `update-article-comment`, `delete-article-comment`).
3. **الكينونة `attachment` (المرفقات)**:
   - رفع وتحديث وحذف الملفات المرفقة بقضايا مشروع معين (`add-attachment`, `update-attachment`, `delete-attachment`).
4. **الكينونة `issueWorkItem` (تتبع الوقت وساعات العمل)**:
   - قراءة وإضافة وتعديل ساعات العمل المجهزة لقضايا مشروع معين (`read-work-item`, `create-work-item`, `update-work-item`, `update-not-own-work-item`, `create-not-own-work-item`).
5. **الكينونة `article` (المقالات وقاعدة المعرفة)**:
   - إدارة وقراءة المقالات التابعة لمشروع معين (`read-article`, `create-article`, `update-article`, `delete-article`).
6. **الكينونة `project` (إدارة وتصفح مشروع قائم)**:
   - الصلاحيات المطبقة على مشروع قائم وموجود بالفعل تكون محددة بنطاق ذلك المشروع (`read-project-basic`, `read-project-full`, `update-project`, `delete-project`).

---

### ج) جدول تصنيفي شامل لكافة الكيانات والصلاحيات الـ 53

| الكيان (`PermissionEntityType`) | اسم الوحدة (`LocalizedModule`) | عدد الصلاحيات | هل هي العامة (`isGlobal`)؟ | أهم الصلاحيات النماذج |
| :--- | :--- | :---: | :---: | :--- |
| **`system`** | System | 2 | **`true` (عامة)** | `low-level-admin-read`, `low-level-admin-write` |
| **`user`** | User | 6 | **`true` (عامة)** | `create-user`, `update-user`, `delete-user`, `read-user-basic`, `read-user-details`, `update-self` |
| **`organization`** | Organization | 4 | **`true` (عامة)** | `create-organization`, `read-organization`, `update-organization`, `delete-organization` |
| **`app`** | App | 2 | **`true` (عامة)** | `read-app-content`, `update-app-content` |
| **`project`** | Project | 5 | **مشترك** (`create-project` فقط `true` والـ 4 الباقية `false`) | `create-project` (عامة) <br> `read-project-basic`, `read-project-full`, `update-project`, `delete-project` (مرتبطة بمشروع) |
| **`issue`** | Issue / Visibility | 12 | **`false` (مرتبطة بمشروع)** | `read-issue`, `create-issue`, `update-issue`, `delete-issue`, `override-visibility-restrictions` |
| **`comment`** | Comment | 10 | **`false` (مرتبطة بمشروع)** | `create-issue-comment`, `read-issue-comment`, `read-article-comment` |
| **`attachment`** | Attachment | 3 | **`false` (مرتبطة بمشروع)** | `add-attachment`, `update-attachment`, `delete-attachment` |
| **`issueWorkItem`** | Issue Work Item | 5 | **`false` (مرتبطة بمشروع)** | `read-work-item`, `create-work-item`, `update-work-item` |
| **`article`** | Article | 4 | **`false` (مرتبطة بمشروع)** | `read-article`, `create-article`, `update-article`, `delete-article` |

---

## 5. ماذا تعني "نوع الكينونة" (Permission Entity Type)؟ وهل هناك مستويات ونطاقات للصلاحيات؟

### أ) معنى "نوع الكينونة" (Permission Entity Type)

نوع الكينونة هو الوسم البرمجي والمعماري (Domain Entity Label) الذي يحدد **المورد المستهدف بالعملية** (Target Resource Type).
فوائده:

1. **التصنيف والتنظيم**: يتيح للواجهات وخدمات الأمان تصنيف الصلاحيات وعرضها للمشرفين في مجموعات منطقية.
2. **الربط مع البرمجيات الوسيطة (Middlewares)**: يتيح لمنطق الأمان معرفة نوع الكائن المراد فحصه (هل افحص الجدول `issues` أم `users` أم `projects`؟).

---

### ب) المستويات والنطاقات الأربعة للأمان في YouTrack

يقسم YouTrack التحقق من الأمان إلى **4 مستويات ونطاقات متدرجة**:

```text
+-----------------------------------------------------------------------+
| 1. النطاق العام (Global Scope)                                         |
|    يسري على النظام ككل (Users, System, Organizations, Create Project) |
+-----------------------------------------------------------------------+
       |
       v
+-----------------------------------------------------------------------+
| 2. نطاق المؤسسة (Organization Scope)                                   |
|    يسري على كافة المشاريع المندرجة تحت مؤسسة معينة                    |
+-----------------------------------------------------------------------+
       |
       v
+-----------------------------------------------------------------------+
| 3. نطاق المشروع (Project Scope)                                        |
|    يسري على مشروع واحد محدد وكافة مخرجاته (Issues, Articles)          |
+-----------------------------------------------------------------------+
       |
       v
+-----------------------------------------------------------------------+
| 4. نطاق العنصر ورؤيته (Resource / Item Level Scope & Visibility)       |
|    قيود الرؤية والحقول الخاصة (Private Fields, Visibility Restrictions)|
+-----------------------------------------------------------------------+
```

1. **المستوى 1: النطاق العام (Global Scope)**
   - الصلاحيات العامة التي تُمنح على مستوى النظام بكامله.
   - المستخدم يملك هذه الصلاحية بغض النظر عن المشروع الذي يعمل عليه.

2. **المستوى 2: نطاق المؤسسة (Organization Scope)**
   - يتيح إسناد أدوار للمستخدمين على مستوى مؤسسة كاملة، مما يمنحهم هذه الأدوار في كافة المشاريع التابعة لتلك المؤسسة تلقائياً.

3. **المستوى 3: نطاق المشروع (Project Scope)**
   - النطاق الأكثر استخداماً في YouTrack. يتم إسناد دور للمستخدم في مشروع محدد (مثل: *Project Developer* في مشروع *Backend*).
   - لا يمتلك المستخدم هذه الصلاحيات في مشاريع أخرى إلا إذا أُسندت له فيها صراحةً.

4. **المستوى 4: نطاق المورد/العنصر القيودي (Resource-Level Scope & Visibility)**
   - صلاحيات دقيقة داخل المشروع، مثل:
     - **الحقول الخاصة (`Read Issue Private Fields`)**: قراءة حقول معينة داخل القضية لا يراها المستخدم العادي.
     - **تجاوز قيود الرؤية (`Override Visibility Restrictions`)**: قراءة قضايا أو تعليقات تم تحديد رؤيتها لمجموعة صريحة فقط.

---

### ج) التدرج وسلاسل الاعتمادية داخل الكيان (Permission Hierarchy)

توجد داخل كل كيان مستويات متدرجة في القوة، ترتبط بسلاسل من `Implies` و `Dependents`.

#### مثال من كيان المشروع (`Project`)

```text
delete-project / update-project  ==[implies]==>  read-project-full  ==[implies]==>  read-project-basic
```

- المستوى الأعلا: `delete-project` أو `update-project` (تعديل وحذف المشروع).
- المستوى المتوسط: `read-project-full` (قراءة تفاصيل المشروع الشاملة).
- المستوى الأساسي: `read-project-basic` (قراءة معلومات المشروع الأساسية).

إذا كان المستخدم يملك `update-project` في مشروع ما، فإن منطق الصلاحيات يستنتج تلقائياً امتلاكه لـ `read-project-full` و `read-project-basic` دون الحاجة لمنحها له بشكل منفصل.

---

## 6. مطابقة الدراسة مع الكود المصدري وقاعدة البيانات في هذا المشروع

تم إثبات ومطابقة هذه المفاهيم داخل مشروعنا الحرفي كالتالي:

### أ) تمثيل كائن الصلاحية في Go (`internal/services/permissions/permissions.go`)

```go
type Permission struct {
 ID                            string     // المعرف التقني kebab-case
 Name                          string     // الاسم للعرض والبحث
 Description                   string     // الوصف الوظيفي
 PermissionEntityType          Entity     // نوع الكينونة (project, issue, user...)
 LocalizedPermissionEntityType Module     // اسم الوحدة للعرض
 IsGlobal                      bool       // تحديد هل الصلاحية عامة أم لا
 Operation                     Operation  // العملية (read, create, update, delete...)
 Dependents                    []string   // الصلاحيات التي تتطلب هذه الصلاحية كشرط
 Implies                       []string   // الصلاحيات الممنوحة ضمنياً مع هذه الصلاحية
}
```

### ب) تمثيل الجداول في PostgreSQL (`migrations/000001_init_schema.up.sql`)

1. **جدول الصلاحيات (`permissions`)**:
   يحوي العمود `is_global BOOLEAN DEFAULT FALSE`.
2. **جدول الأدوار المُسندة (`assigned_roles`)**:
   يحوي الأعمدة `scope_type`, `scope_id`, `scope_project_id`, `scope_organization_id` لتمثيل نطاقات الإسناد (Global / Project / Organization).
3. **جدول ذاكرة الصلاحيات المؤقتة (`cached_permissions`)**:
   يحوي العمود `is_global BOOLEAN DEFAULT FALSE` لكل مستخدم وصلاحية.

### ج) البذرة البياناتية (`migrations/000016_rbac_seed.up.sql`)

تتضمن إدخال الصلاحيات الـ 53 الكاملة مع تحديد قيم `is_global` بدقة مطابق لما ورد في هذه الدراسة:

- `is_global = TRUE` لجميع صلاحيات `User`, `System`, `Organization`, `App`, وصلاحية `create-project`.
- `is_global = FALSE` لجميع صلاحيات `Issue`, `Comment`, `Attachment`, `IssueWorkItem`, `Article`, وصلاحيات قراءة/تعديل/حذف مشروع قائم.

---

## 7. الخلاصة التنفيذية

1. **نظام RBAC ونطاقاته**: الصلاحيات في YouTrack تجمعها أدوار وتُسند ضمن نطاقات (Global, Organization, Project).
2. **خاصية `global`**: تعني أن الصلاحية تعمل على مستوى النظام ككل، وتتطلب إسناداً في النطاق العام (Global Scope).
3. **فحص الأمان لـ `global`**: **مطلوب إجبارياً**، ولا تعني `global` إمكانية الوصول بدون صلاحية، بل تعني أن الفحص يتم على النطاق العام للنظام.
4. **الكينات العامة**: تشمل إدارة النظام (`system`)، الحسابات (`user`)، المؤسسات (`organization`)، التطبيقات (`app`)، وإنشاء المشروعات (`create-project`).
5. **مستويات الصلاحيات**: تتوزع الأذونات على 4 نطاقات أمان (Global, Org, Project, Item/Visibility) مع وجود تدرج هرمي وسلاسل استلزام (`Implies`/`Dependents`) تتيح منح الصلاحيات الأضيق تلقائياً عند منح الصلاحيات الأوسع.
