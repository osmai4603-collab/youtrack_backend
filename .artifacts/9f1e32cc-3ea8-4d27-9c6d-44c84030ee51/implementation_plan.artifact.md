# إعداد API Endpoint للطلب رقم 30 (securitySearch/filterFields)

يتمثل الهدف في تنفيذ نقطة نهاية `/api/securitySearch/filterFields` التي تعيد حقول التصفية الخاصة بالأمان بناءً على نوع الكيان (`entityType`). سنقوم بإنشاء نموذج بيانات جديد وجدول قاعدة بيانات جديد (Scheme) لضمان عدم تعديل المخططات الحالية، مع اتباع نمط جلب البيانات المستخدم في طلبات "المستخدم الحالي" لدعم معامل `fields`.

## Proposed Changes

### [Database Layer]

#### [NEW] [000014_security_filter_fields.up.sql](file:///home/osmsoftwareengineering/StudioProjects/youtrack_backend/db/migrations/000014_security_filter_fields.up.sql)
إنشاء جدول جديد `security_filter_fields` لتخزين حقول تصفية الأمان.
- الأعمدة: `id`, `name`, `entity_type`, `field_type`.
- إدراج البيانات الأولية لـ `ProjectPeopleResponse` (Scope, Role, Permission).

### [Model Layer]

#### [NEW] [security_search.go](file:///home/osmsoftwareengineering/StudioProjects/youtrack_backend/internal/model/security_search.go)
تعريف هيكل `SecurityFilterField` لتمثيل حقول التصفية في استجابة JSON.

### [Store Layer]

#### [MODIFY] [store.go](file:///home/osmsoftwareengineering/StudioProjects/youtrack_backend/internal/store/store.go)
إضافة واجهة `SecuritySearchStore` للتعامل مع حقول تصفية الأمان.

#### [NEW] [security_search_store.go](file:///home/osmsoftwareengineering/StudioProjects/youtrack_backend/internal/store/sqlstore/security_search_store.go)
تنفيذ واجهة `SecuritySearchStore` باستخدام PostgreSQL لجلب البيانات من الجدول الجديد.

### [App Layer]

#### [MODIFY] [app.go](file:///home/osmsoftwareengineering/StudioProjects/youtrack_backend/internal/app/app.go)
إضافة منطق العمل (Business Logic) لجلب حقول تصفية الأمان وتنسيقها.

### [API Layer]

#### [NEW] [security_search.go](file:///home/osmsoftwareengineering/StudioProjects/youtrack_backend/internal/api/security_search.go)
إنشاء معالج (Handler) لطلب `GET /api/securitySearch/filterFields`.
- دعم معامل `fields` للاختيار الديناميكي للحقول.
- دعم معامل `entityType` للتصفية.

#### [MODIFY] [router.go](file:///home/osmsoftwareengineering/StudioProjects/youtrack_backend/internal/api/router.go)
تسجيل المسار الجديد في الـ Router.

## Verification Plan

### Automated Tests
- إنشاء اختبار وحدة في `internal/api/security_search_test.go` للتحقق من:
  - جلب الحقول الصحيحة لـ `ProjectPeopleResponse`.
  - احترام معامل `fields`.
  - التعامل مع أنواع الكيانات غير الموجودة.

### Manual Verification
- استخدام `curl` لطلب نقطة النهاية والتأكد من مطابقة الاستجابة لـ `request30.txt`.
