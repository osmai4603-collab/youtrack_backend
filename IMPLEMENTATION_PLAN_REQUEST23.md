# خطة تنفيذ API Endpoint للطلب رقم 23 (GET /api/admin/projects/{id}/dashboard)

## 1. ملخص الحالة الحالية (Audit Result)

**البنية الرأسية (vertical slice) للنقطة موجودة بالفعل بالكامل:**

| الطبقة | الملف | الحالة |
|--------|-------|--------|
| الجدول `project_dashboard_widgets` | `db/migrations/000001_init_schema.up.sql` | ✅ موجود |
| الجدول `dashboard_widgets` | `db/migrations/000001_init_schema.up.sql` | ✅ موجود |
| النموذج `ProjectDashboard`/`ProjectDashboardWidget` | `internal/model/dashboard.go` | ✅ موجود |
| واجهة الـ Store | `internal/store/store.go:90` | ✅ موجودة |
| تنفيذ الـ Store | `internal/store/sqlstore/admin_store.go:297` | ✅ موجود |
| طبقة الـ App | `internal/app/admin.go:61` | ✅ موجودة |
| الـ Handler + serializer | `internal/api/admin.go:303` | ✅ موجود |
| المسار | `internal/api/router.go:62` | ✅ موجود |
| الاختبارات | `internal/api/admin_test.go` | ✅ موجودة |

**لا حاجة لإنشاء ملفات نماذج/Store/Handler جديدة** — يجب **تعديل النماذج الموجودة** و**إصلاح قيم `$type`** وإضافة **بذر البيانات** فقط.

---

## 2. الفجوة الحرجة (Critical Discrepancy)

مقارنة استجابة `docs/requests/request23.txt` مع ما ينتجه الكود الحالي:

| العنصر | الكود الحالي | المطلوب في request23.txt |
|--------|--------------|--------------------------|
| `$type` للودجت المدمجة | `"ProjectDashboardWidget"` | **`"DashboardWidgetEmbedding"`** |
| `$type` للودجت الفرعي المتداخل | `"WidgetView"` | **`"EmbeddingWidget"`** |

ملاحظة: طلب `request13.txt` يعيد `{"widgets":[],"$type":"ProjectDashboard"}` (مصفوفة فارغة)، لذا لم يُفحص `$type` الداخلي من قبل. request23 يكشف القيم الحقيقية.

**الفجوة الثانية:** لا يوجد بذر لصفوف `project_dashboard_widgets` للمشروع (`0-0`) — المشروع `0-0` مذكور في السيدر (agile boards، إلخ) لكن لا صفوف للوحة في `project_dashboard_widgets`.

---

## 3. التغييرات المقترحة (Proposed Changes)

### 3.1 تصحيح قيم `$type` في النموذج — `internal/model/dashboard.go`

تعديل دالة `Normalize` في `ProjectDashboardWidget` لتعكس المخطط الحقيقي:

```go
// Normalize يملأ $type للودجت المدمجة (مطابق لـ request23.txt).
func (w *ProjectDashboardWidget) Normalize() {
	w.Type = "DashboardWidgetEmbedding"
	if w.Widget != nil && w.Widget.Type == "" {
		w.Widget.Type = "EmbeddingWidget"
	}
}
```

### 3.2 تصحيح قيم `$type` في متسلسل الـ API — `internal/api/admin.go`

تعديل دالتين:

```go
// projectDashboardWidgetToMap — تغيير السطر الأخير:
result["$type"] = "DashboardWidgetEmbedding"  // كانت "ProjectDashboardWidget"

// projectDashboardWidgetViewToMap — تغيير السطر الأخير:
result["$type"] = "EmbeddingWidget"  // كانت "WidgetView"
```

### 3.3 محاذاة قيمة `$type` في الـ Store — `internal/store/sqlstore/admin_store.go`

في `ProjectDashboard` تُضبط `w.Type` يدوياً على `"ProjectDashboardWidget"` (سطر 333). **لتكون القيم متطابقة عبر جميع المسارات وتنتهي بـ `Normalize()` من طبقة الـ App، يمكن تركه** لأنه يتم استبداله بـ `Normalize()` في `GetProjectDashboard`. لكن إبقاءه متسقاً (أو إفراغه وترك `Normalize` يملؤه) هو الأنظف:
- الخيار الموصى به: إبقاء `Normalize()` كالمصدر الوحيد للحقيقة وترك قيمة stotre كما هي (تُستبدل لاحقاً). هذا لا يسبب مشكلة لأن `GetProjectDashboard` يستدعي `dashboard.Normalize()` والذي يستدعي `Normalize()` لكل ودجت.

**قرار:** بما أن `DashboardWidget.Type` مملوء بـ `"WidgetView"` في الـ Store (سطر 353) ويكون عندها `w.Widget.Type != ""`، **لن يستبدل** `Normalize()` القيمة تلقائياً. الحل الآمن: تعديل الـ Store ومتسلسل الـ API والنموذج معاً، أو إفراغ `Widget.Type` في الـ Store وأيضاً ضبط `Type` في النموذج. **الأسلوب الموصى به (مزدوج التوثيق):**

في `admin_store.go` السطر 333: `w := &model.ProjectDashboardWidget{}` (بدون ضبط Type يدوي — يملؤه Normalize).
السطر 353: `w.Widget = &model.DashboardWidget{ID: ""}` (بدون ضبط Type — يملؤه Normalize بالـ EmbeddingWidget).

في `GetProjectDashboard` (`internal/app/admin.go:69`): يستدعي `dashboard.Normalize()` الذي يملأ `$type` لكل ودجت (سطر 29-33 من النموذج). **لكن** لا يملأ `$type` للودجت الفرعي إلا إذا كان `w.Widget.Type == ""`. لذا يجب التأكد أن الـ Store لا يضبط `Widget.Type`.

> **الخلاصة:** نزيل ضبط `Type` اليدوي من الـ Store ونترك `Normalize()` في النموذج (بعد تعديله) هو المصدر الوحيد للقيم الصحيحة، ويبقى متسلسل الـ API خارجياً هو الذي يُخرج `$type` في JSON (فهو لا يعتمد على `w.Type` بل يكتبه بنفسه).

**النصيحة لتجنب أي التباس:** بما أن متسلسل `projectDashboardWidgetToMap`/`projectDashboardWidgetViewToMap` **يكتب `$type` بنفسه في الخريطة** (لا يقرأ من `w.Type`)، فإن تعديل هذين السطرين في `admin.go` **كافٍ تماماً** لإخراج `$type` الصحيح في JSON. تعديل النموذج و `Normalize` اختياري للنظافة/الاكتمال.

### 3.4 بذر بيانات المشروع `0-0` — هجرة جديدة

جدول `project_dashboard_widgets` لا يحتوي صفوفاً. إضافة هجرة:

**[NEW] `db/migrations/000013_project_dashboard_widgets_seed.up.sql`**

```sql
-- 000013_project_dashboard_widgets_seed.up.sql
-- Request #23: GET /api/admin/projects/0-0/dashboard
-- يلقّح ودجات لوحة المشروع 0-0 (مطابقة لـ request23.txt).

-- ضمان وجود الودجات المرجعية في dashboard_widgets (174-7, 174-3)
INSERT INTO dashboard_widgets (id, key)
VALUES
    ('174-7', 'hbr'),
    ('174-3', 'mob')
ON CONFLICT (id) DO NOTHING;

-- ضمان وجود المشروع 0-0 (اسمه "Demo project")
INSERT INTO projects (id, name, short_name, is_demo)
VALUES ('0-0', 'Demo project', '0-0', TRUE)
ON CONFLICT (id) DO NOTHING;

-- بذر صفوف اللوحة المدمجة
INSERT INTO project_dashboard_widgets (id, project_id, widget_id, key, x, y, width, height, settings)
VALUES
    ('178-4', '0-0', '174-7', 'hbr', 0, 0, 2, 2,
     '{"customWidgetConfig":"{ \"selectedProject\": { \"key\": \"0-0\", \"label\": \"Demo project\" }, \"title\": \"Project Team — Demo project\" }"}'),
    ('178-5', '0-0', '174-3', 'mob', 2, 0, 2, 2,
     '{"customWidgetConfig":"{ \"search\": \"project: {Demo project} State: Unresolved\", \"title\": \"Issue List — Unresolved in Demo project\" }"}')
ON CONFLICT (id) DO UPDATE SET
    project_id = EXCLUDED.project_id,
    widget_id  = EXCLUDED.widget_id,
    key        = EXCLUDED.key,
    x          = EXCLUDED.x,
    y          = EXCLUDED.y,
    width      = EXCLUDED.width,
    height     = EXCLUDED.height,
    settings   = EXCLUDED.settings;
```

**[NEW] `db/migrations/000013_project_dashboard_widgets_seed.down.sql`**

```sql
-- 000013_project_dashboard_widgets_seed.down.sql
DELETE FROM project_dashboard_widgets WHERE project_id = '0-0' AND id IN ('178-4', '178-5');
```

### 3.5 (اختياري) تمديد الـ Seeder — `cmd/seeder/main.go`

للحفاظ على تناسق عملية البذر، إضافة `scanAndSeedProjectDashboardWidgets` تقرأ `request23.txt` وتلقّح `project_dashboard_widgets` و `dashboard_widgets` (174-7, 174-3). **يُوصى به** لأنه نمط باقي السيدر. يمكن تنفيذه بالتوازي مع الهجرة (كلاهما idempotent بـ `ON CONFLICT`).

### 3.6 الاختبارات — `internal/api/admin_test.go`

1. تحديث دالة `ProjectDashboard` في `mockAdminStore` لمحاكاة مشروع `0-0` بصفين يتطابقان مع request23 (hbr, mob) — مع استخدام النظام الجديد لبناء البيانات.
2. إضافة `TestProjectDashboardRequest23` — يتحقق من المطابقة الدقيقة للاستجابة (كل الحقول، القيم، و `$type` = `DashboardWidgetEmbedding`/`EmbeddingWidget`).
3. إبقاء `TestProjectDashboardFieldFiltering` وإضافة فحص أن `widget(id)` يعمل مع `widget` المتداخل.
4. **تحديث** الحالات الموجودة (`TestProjectDashboardRequest13`, `TestProjectDashboardFullSchema`) التي تفحص `$type == "ProjectDashboardWidget"`/`"WidgetView"` إلى القيم الجديدة `"DashboardWidgetEmbedding"`/`"EmbeddingWidget"`.
5. إبقاء `TestProjectDashboardNotFound` (404) و`TestProjectDashboardUnauthorized` (401).

---

## 4. خطة التحقق (Verification Plan)

### التجميع والاختبارات:
```bash
go build ./...
go test -v ./...
go test -v ./internal/api -run TestProjectDashboard
```

### فحص النقطة يدوياً:
1. تشغيل الـ seeder: `go run ./cmd/seeder`
2. تشغيل الهجرة (الجديدة 000013 + القائمة).
3. تشغيل الخادم واستدعاء:
   ```
   GET /api/admin/projects/0-0/dashboard?fields=widgets(id,key,x,y,width,height,widget(id),settings)
   ```
   ومقارنة الناتج مع `docs/requests/request23.txt` حرفياً.

---

## 5. ملخص الملفات المتأثرة

| الملف | نوع التغيير |
|-------|-------------|
| `internal/model/dashboard.go` | تعديل: قيم `$type` في `Normalize` |
| `internal/api/admin.go` | تعديل: `$type` في `projectDashboardWidgetToMap` و `projectDashboardWidgetViewToMap` |
| `internal/store/sqlstore/admin_store.go` | تعديل (اختياري): إزالة ضبط `Type` اليدوي ليعتمد على `Normalize` |
| `internal/api/admin_test.go` | تعديل: mock + تحديث الاختبارات + إضافة `TestProjectDashboardRequest23` |
| `db/migrations/000013_project_dashboard_widgets_seed.up.sql` | جديد: بذر |
| `db/migrations/000013_project_dashboard_widgets_seed.down.sql` | جديد: تراجع |
| `cmd/seeder/main.go` | تعديل (اختياري): `scanAndSeedProjectDashboardWidgets` |
