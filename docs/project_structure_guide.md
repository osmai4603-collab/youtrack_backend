# الدليل الشامل والمفصل لهيكلية وملفات مشروع YouTrack Backend

يحتوي هذا المستند على شرح دقيق وتفصيلي لكافة المجلدات، والملفات، والنماذج، ونقاط النهاية (API Endpoints)، والآليات التقنية المتبعة في مشروع **YouTrack Backend**.

---

## 1. شجرة الملفات العامة للمشروع

```text
youtrack_backend/
├── cmd/
│   └── api/
│       └── main.go                 # نقطة الدخول الرئيسية، تسجيل المسارات، والـ Graceful Shutdown
├── internal/
│   ├── config/
│   │   └── config.go               # قراءة المتغيرات البيئية وإعدادات تشغيل الخادم
│   ├── domain/
│   │   └── models.go               # نماذج وهياكل البيانات الأساسية ومطابقة مواصفات YouTrack REST API
│   ├── handler/                    # طبقة المعالجات واستقبال طلبات الـ HTTP
│   │   ├── admin.go                # معالجات إعدادات النظام، ساعات العمل، والـ Widgets
│   │   ├── config.go               # معالجات إعدادات واجهة المستخدم والبانرات
│   │   ├── health.go               # فحص جاهزية وصحة الخدمة (Health Check)
│   │   ├── issue.go                # معالجات أنواع الروابط وسجل الأنشطة (Activities)
│   │   ├── permission.go           # معالجات الصلاحيات المخزنة للمستخدم (Permissions Cache)
│   │   └── user.go                 # معالجات الملف الشخصي والتفضيلات ومساعد Grazie
│   └── middleware/
│       └── logger.go               # وسيط تسجيل الطلبات وتوقيت الاستجابة
├── docs/
│   ├── architecture.md             # التوثيق المعماري العام ومخطط تدفق البيانات
│   └── project_structure_guide.md  # هذا الدليل التفصيلي الشامل
├── .env.example                    # نموذج لمتغيرات البيئة اللازمة للتشغيل
├── .gitignore                      # استثناء ملفات البناء والبيئة من Git
├── Makefile                        # أوامر الأتمتة السريعة (بناء، تشغيل، اختبار، تنظيف)
└── go.mod                          # تعريف الموديول وإصدار لغة Go
```

---

## 2. النمط المعماري للمشروع (Clean Architecture)

يعتمد المشروع على المعمارية النظيفة مع تقسيم **Standard Go Project Layout** لضمان:
1. **عزل منطق النطاق (Domain Isolation)**: عدم اعتماد نماذج العمل على أي مكتبات خارجية أو بروتوكولات نقل.
2. **الاستقلالية وقابلية التوسع (Scalability & Decoupling)**: إمكانية استبدال أو تحديث طبقة الـ HTTP أو إضافة طبقة قاعدة بيانات دون التأثير على العقود والنماذج.
3. **أداء عالٍ بأقل اعتماديات (Zero Third-Party Dependencies)**: الاعتماد التام على مكتبة Go القياسية (`net/http`, `encoding/json`, `os/signal`).

---

## 3. شرح تفصيلي للملفات والمجلدات

### 1️⃣ مجلد نقطة البداية `cmd/api/`

- **[`cmd/api/main.go`](file:///home/osmsoftwareengineering/StudioProjects/youtrack_backend/cmd/api/main.go)**:
  - **الدور**: نقطة انطلاق التطبيق (`func main()`).
  - **الوظائف**:
    - تحميل الإعدادات عبر `config.LoadConfig()`.
    - تهيئة موجه المسارات `http.NewServeMux()` باستخدام ميزات توجيه Go 1.22+ المدمجة (مثل `GET /api/issues/{issueID}/activitiesPage`).
    - تسجيل مسارات Health Check، والـ Config، والمستخدمين، والصلاحيات، والمهام، ولوحة الإدارة.
    - تطبيق وسيط التسجيل `middleware.Logger(mux)`.
    - ضبط الـ Timeouts للخادم (`ReadTimeout: 10s`, `WriteTimeout: 10s`, `IdleTimeout: 60s`).
    - إدارة الإيقاف السلس (**Graceful Shutdown**) باستخدام قنوات الـ OS Signal (`SIGINT`, `SIGTERM`) مع مهلة 5 ثوانٍ لإنهاء الطلبات المعلقة دون قطع مفاجئ.

---

### 2️⃣ مجلد التهيئة والإعدادات `internal/config/`

- **[`internal/config/config.go`](file:///home/osmsoftwareengineering/StudioProjects/youtrack_backend/internal/config/config.go)**:
  - **الدور**: إدارة إعدادات تشغيل التطبيق.
  - **الهياكل**:
    - `Config`: يحمل حقول `Port` (الافتراضي: `8080`) و `AppEnv` (الافتراضي: `development`).
  - **الدوال**:
    - `LoadConfig()`: يقرأ المتغيرات البيئية `SERVER_PORT` و `APP_ENV` مع توفير قيم افتراضية آمنة في حال غيابها.

---

### 3️⃣ مجلد النطاق والنماذج `internal/domain/`

- **[`internal/domain/models.go`](file:///home/osmsoftwareengineering/StudioProjects/youtrack_backend/internal/domain/models.go)**:
  - **الدور**: يحتوي على كافة الهياكل المعيارية المتوافقة بنسبة 100% مع واجهات YouTrack REST API، بما في ذلك وسوم النوع `$type` المستخدمة في JetBrains YouTrack.
  - **أبرز أقسام النماذج**:
    1. **المستخدمين والمجموعات (Users & Groups)**:
       - `User`: البيانات الشخصية، الحالة (`Online`, `Banned`, `IsLocked`)، الملفات التعريفية والميزات التجريبية.
       - `UserType`, `UserGroup`, `ProjectTeam`, `Visibility`.
    2. **المشاريع والمهام (Projects & Issues)**:
       - `Project`: تفاصيل المشروع وحالته (`Archived`, `Pinned`, `Restricted`).
       - `Issue`: التذكرة أو المهمة، وتتضمن الحقول المخصصة، المرفقات، التعليقات، الروابط، والمصوتين.
       - `IssueCustomField`, `FieldDefinition`, `CustomFieldType`, `ProjectCustomField`.
       - `IssueLinkType`, `IssueLink`: إدارة علاقات المهام (`Subtask`, `Relates`, `Duplicate`).
    3. **المرفقات والتعليقات (Attachments & Comments)**:
       - `Attachment`: الملفات، الأنواع `MimeType`، الأبعاد `ImageDimensions`، والنصوص المستخرجة ضوئياً `RecognizedText`.
       - `Comment`: نص التعليق، المرفقات، والتفاعلات بالرموز التعبيرية `Reaction`.
    4. **سجل الأنشطة والترقيم بالمؤشرات (Activities & Cursor Pagination)**:
       - `ActivityItem`: يمثل الحدث (إضافة تذكرة، تعديل حقل، إضافة تعليق) مع المؤلف والهدف والحقول المتأثرة.
       - `ActivityCursorPage`: يدعم التصفح المعتمد على المؤشرات (`Cursor`, `BeforeCursor`, `AfterCursor`, `HasBefore`, `HasAfter`).
    5. **ملفات التفضيلات الشخصية (User Profiles & Preferences)**:
       - `UserProfiles`: يجمع كافة ملفات التخصيص:
         - `GeneralUserProfile`: المنطقة الزمنية، صيغة التاريخ، اللغة.
         - `AppearanceUserProfile`: تفضيلات الثيم، عرض الجداول، القوائم الجانبية، ترتيب التعليقات.
         - `IssuesListUserProfile`: إعدادات عرض شجرة وتصفية قائمة المهام.
         - `ArticlesUserProfile`: إعدادات مقالات قاعدة المعرفة.
         - `NotificationsUserProfile`: تفضيلات استلام إشعارات البريد الإلكتروني وتتبع المهام التلقائي.
         - `TimeTrackingUserProfile`: صيغة تسجيل الوقت ومخططات الساعات.
         - `HelpdeskUserProfile`: صلاحيات ومشاريع مكتب المساعدة.
         - `TipsUserProfile`: حفظ حالة ظهور النوافذ الترحيبية والتلميحات والإرشادات.
         - `GrazieUserProfile` & `AiUserProfile`: إعدادات مساعد الكتابة الذكي والتدقيق اللغوي وحجم نافذة الشات.
    6. **إعدادات النظام والإدارة (System & Admin Settings)**:
       - `WorkTimeSettings`: أيام العمل الأسبوعية، الدقائق اليومية، وأول أيام الأسبوع.
       - `GlobalSettings`: إعدادات الـ OCR، خادم البريد `NotificationSettings`، وإعدادات الـ REST/CORS.
       - `WidgetView`: مواصفات أدوات لوحة التحكم والمقالات وتفاصيل المطور والأبعاد الافتراضية.
       - `CachedPermission`: كاش الصلاحيات العامة وعلى مستوى المشاريع (`JetPass` / `YouTrack`).

---

### 4️⃣ مجلد المعالجات `internal/handler/`

- **[`internal/handler/health.go`](file:///home/osmsoftwareengineering/StudioProjects/youtrack_backend/internal/handler/health.go)**:
  - معالجة طلبات فحص الصحة على `/health` و `/api/v1/health`.
  - يرجع حالة `"OK"` وتوقيت UTC واسم الخدمة.

- **[`internal/handler/config.go`](file:///home/osmsoftwareengineering/StudioProjects/youtrack_backend/internal/handler/config.go)**:
  - معالجة طلب `/api/config`.
  - يرجع إعدادات واجهة YouTrack وحالة تفعيل مكتب المساعدة وإعلانات النظام العامة `BannersConfig`.

- **[`internal/handler/user.go`](file:///home/osmsoftwareengineering/StudioProjects/youtrack_backend/internal/handler/user.go)**:
  - `GetCurrentUser`: معالجة `/api/users/me` وإرجاع تفاصيل المستخدم المسجل مع جميع ملفات التفضيلات المتكاملة.
  - `GetGrazieProfile`: معالجة `/api/users/me/profiles/grazie` الخاصة بإعدادات مساعد الذكاء الاصطناعي والتدقيق الإملائي.

- **[`internal/handler/permission.go`](file:///home/osmsoftwareengineering/StudioProjects/youtrack_backend/internal/handler/permission.go)**:
  - معالجة `/api/permissions/cache`.
  - يرجع قائمة الصلاحيات المفعلة للمستخدم الحالي للمشاريع الافتراضية ومشاريع الـ Helpdesk (مثل `READ_ISSUE`, `CREATE_ISSUE`, `CREATE_COMMENT`, `user-read-basic`).

- **[`internal/handler/issue.go`](file:///home/osmsoftwareengineering/StudioProjects/youtrack_backend/internal/handler/issue.go)**:
  - `GetIssueLinkTypes`: معالجة `/api/issueLinkTypes` لإرجاع أنواع العلاقات بين المهام (Subtask, Relates, Duplicate).
  - `GetActivitiesPage`: معالجة `/api/issues/{issueID}/activitiesPage` واستخراج `issueID` ديناميكياً وإرجاع سجل الأنشطة والترقيم بالمؤشر (Cursor Pagination).

- **[`internal/handler/admin.go`](file:///home/osmsoftwareengineering/StudioProjects/youtrack_backend/internal/handler/admin.go)**:
  - `GetWorkTimeSettings`: معالجة `/api/admin/timeTrackingSettings/workTimeSettings` (ساعات العمل وأيام الأسبوع).
  - `GetGlobalSettings`: معالجة `/api/admin/globalSettings` (CORS، الإشعارات، التعرف الضوئي OCR).
  - `GetGeneralWidgets`: معالجة `/api/admin/widgets/general` وإرجاع الـ Widgets الافتراضية للوحة التحكم.

---

### 5️⃣ مجلد الوسائط `internal/middleware/`

- **[`internal/middleware/logger.go`](file:///home/osmsoftwareengineering/StudioProjects/youtrack_backend/internal/middleware/logger.go)**:
  - وسيط (HTTP Middleware) يقوم باعتراض كل طلب وارد وحساب زمن المعالجة وتسجيل نوع الطلب والمسار:
    ```log
    [GET] /api/users/me 120.45µs
    ```

---

### 6️⃣ الملفات الجذرية وملفات الأتمتة

- **[`go.mod`](file:///home/osmsoftwareengineering/StudioProjects/youtrack_backend/go.mod)**:
  - تعريف الموديول `youtrack_backend` واستخدام Go 1.26 (أو 1.22+ للاستفادة من قدرات الـ Mux المدمجة).
- **[`Makefile`](file:///home/osmsoftwareengineering/StudioProjects/youtrack_backend/Makefile)**:
  - أوامر التشغيل السريعة:
    - `make build`: بناء الملف الثنائي في `bin/youtrack_backend`.
    - `make run`: تشغيل المشروع محلياً مباشرة.
    - `make test`: تشغيل اختبارات الوحدة مع كاشف السباق (`-race`).
    - `make clean`: إزالة مجلد البناء `bin/`.
    - `make fmt`: تنسيق الكود وفق معايير Go.
    - `make tidy`: تنظيف وترتيب حزم go.mod.
- **[`.env.example`](file:///home/osmsoftwareengineering/StudioProjects/youtrack_backend/.env.example)**:
  - قالب توضيحي للمتغيرات البيئية الخاصة بمنفذ الخادم، والبيئة، والاتصال المستقبلي بقاعدة البيانات وJWT.
- **[`.gitignore`](file:///home/osmsoftwareengineering/StudioProjects/youtrack_backend/.gitignore)**:
  - استبعاد ملفات النظام، وملفات الإعدادات المحلية `.env`، والملفات المبنية الثنائية.

---

## 4. جدول نقاط النهاية المتاحة (API Endpoints Matrix)

| المسار (Route) | الطريقة (Method) | المعالج (Handler) | الوصف |
| :--- | :--- | :--- | :--- |
| `/health` | `GET` | `HealthHandler.Check` | فحص صحة الخادم العامة |
| `/api/v1/health` | `GET` | `HealthHandler.Check` | فحص صحة الخادم (مسبوق بإصدار API) |
| `/api/config` | `GET` | `ConfigHandler.GetConfig` | إعدادات YouTrack العامة والبانرات |
| `/api/users/me` | `GET` | `UserHandler.GetCurrentUser` | تفاصيل المستخدم الحالي وكافة ملفات التفضيلات |
| `/api/users/me/profiles/grazie` | `GET` | `UserHandler.GetGrazieProfile` | إعدادات مساعد الذكاء الاصطناعي Grazie |
| `/api/permissions/cache` | `GET` | `PermissionHandler.GetPermissionsCache` | الصلاحيات المخزنة مؤقتاً للمستخدم |
| `/api/issueLinkTypes` | `GET` | `IssueHandler.GetIssueLinkTypes` | أنواع علاقات وروابط المهام |
| `/api/issues/{issueID}/activitiesPage` | `GET` | `IssueHandler.GetActivitiesPage` | سجل أحداث وتغييرات التذكرة مع الترقيم بالمؤشر |
| `/api/admin/timeTrackingSettings/workTimeSettings` | `GET` | `AdminHandler.GetWorkTimeSettings` | إعدادات وساعات العمل الأسبوعية واليومية |
| `/api/admin/globalSettings` | `GET` | `AdminHandler.GetGlobalSettings` | الإعدادات العامة للنظام (CORS / OCR / البريد) |
| `/api/admin/widgets/general` | `GET` | `AdminHandler.GetGeneralWidgets` | قائمة أدوات لوحة التحكم الافتراضية |

---

## 5. خطة التوسع المستقبلية (Future Roadmap)

عند الرغبة في ربط التطبيق بقاعدة بيانات حقيقية (مثل PostgreSQL)، يتم اتباع المخطط التالي دون كسر المعمارية الحالية:
1. **طبقة المستودعات (`internal/repository/`)**: إضافة مستودعات مثل `user_repo.go`, `issue_repo.go` لتنفيذ عمليات SQL.
2. **طبقة الخدمات (`internal/service/`)**: نقل منطق الحسابات والتحقق من الصلاحيات من الـ Handlers إلى الـ Services.
3. **حقن الاعتماديات في `cmd/api/main.go`**: تمرير الـ Repositories إلى الـ Services ثم تمرير الـ Services إلى الـ Handlers.
