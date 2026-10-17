# تحليل سبب خطأ HTTP 403 Forbidden عند جلب صلاحيات المستخدم (`GET /api/users/me/permissions`)

## 1. ملخص المشكلة

عند تسجيل الدخول بنجاح عبر `/api/auth/login` وإرسال طلب مباشر لجلب صلاحيات المستخدم عبر `/api/users/me/permissions`، يُرجع النظام الاستجابة:
`HTTP/1.1 403 Forbidden` بحمولة خطأ تفيد برفض التخويل (`permissions denied / not authorized`).

---

## 2. السبب المباشر (Direct Cause)

السبب المباشر لرفض الطلب وإرجاع الحالة **403 Forbidden** هو أن المسار `/api/users/me/permissions` مجهّز ببرمجية وسيطة للتخويل (Authorization Middleware) تتطلب توفر الصلاحية العامة **`read-user-basic`** (`perms.PermUserReadUserBasic`).

عندما لا يملك المستخدم الذي سجل دخوله أي دور مُسند في جدول `assigned_roles` على المستوى العام (`scope_type = 'global'`) يمنحه هذه الصلاحية، تفشل دالة فحص التخويل (`svc.IsAllowed`) ويقوم الـ Middleware برفض الطلب وإرجاع 403.

---

## 3. التتبع التفصيلي لمسار التنفيذ في الشفرة البرمجية (Execution Flow Trace)

### أ. تعريف المسار وحمايته (Routing)

* **الملف**: `internal/httphandlers/handles/users/init.go`
* **الكود**:

  ```go
  permissionRequired := func(permission perms.Permission, hlrFunc http.HandlerFunc) http.HandlerFunc {
      return middlewares.RequireGlobal(svc, permission, hlrFunc)
  }

  router := parent.Route("/api/users", func(api chi.Router) {
      api.Use(protector)
      api.Get("/me", permissionRequired(perms.PermUserReadUserDetails, handler.handleMe))
      api.Get("/me/permissions", permissionRequired(perms.PermUserReadUserBasic, handler.handleUserPermissions))
  })
  ```

* **الشرح**: تم ربط المسار `/api/users/me/permissions` بـ `permissionRequired` والتي تستدعي بدورها `middlewares.RequireGlobal(svc, perms.PermUserReadUserBasic, ...)`.

### ب. برمجية فحص التخويل الوسيطة (Authorization Middleware)

* **الملف**: `internal/httphandlers/middlewares/authorize.go`
* **الكود**:

  ```go
  isAllowed, err := svc.IsAllowed(r.Context(), userID, permission, scope)
  if err != nil {
      response.Error(w, r, platErrors.Forbidden(opAuthorize, "permissions unexpected error", err))
      return
  }
  if !isAllowed {
      response.Error(w, r, platErrors.Forbidden(opAuthorize, "permissions denied", errors.New("not authorized")))
      return
  }
  ```

* **الشرح**: تقوم البرمجية الوسيطة بانتزاع معرّف المستخدم (`userID`) من سياق الطلب (Context) وتستدعي `svc.IsAllowed` للتحقق من امتلاكه للصلاحية `PermUserReadUserBasic` بالنطاق العام `ScopeGlobal`.

### ج. الاستعلام من قاعدة البيانات (Storage Query)

* **الملف**: `internal/storage/permissions/get_permissions_for_user.go`
* **الكود**:

  ```sql
  SELECT DISTINCT ar.scope_type AS role_scope, perm
  FROM assigned_roles ar
  JOIN roles r ON r.id = ar.role_id
  CROSS JOIN LATERAL unnest(r.permissions) AS perm
  WHERE (
      (ar.holder_type = 'user' AND ar.holder_id = $1)
      OR (ar.holder_type = 'group' AND ar.holder_id IN (
          SELECT group_id FROM group_members WHERE user_id = $1
      ))
  )
  AND (
      ar.scope_type = 'global'
      OR ar.scope_type IS NULL
      OR ar.scope_type = ''
      ...
  )
  ```

* **الشرح**: تقوم الدالة `GetPermissionsForUser` بالاستعلام في جدول `assigned_roles` والأدوار المرتبطة بها لمستخدم معين أو المجموعات التي ينتمي إليها. إذا لم تجد أي دور مُسند للمستخدم في `assigned_roles` بنطاق عام يتضمن الصلاحية `read-user-basic` (أو ما يكافئها)، فإن قائمة الصلاحيات ترجع فارغة.

### د. تقييم النتيجة

* تنفذ الدالة `HasPermissionGranted` فحصاً مطبقاً على القائمة الخالية أو الخالية من الصلاحية، فترجع `false`.
* يُرجع الـ Middleware الاستجابة `403 Forbidden` وتتوقف معالجة الطلب قبل الوصول إلى دالة المعالج `handler.handleUserPermissions`.

---

## 4. الحالات والسياقات الشائعة للمشكلة (Root Cause Scenarios)

1. **دخول مستخدم عادي / ضيف لا يملك أية أدوار عامة (Non-Admin / Guest User)**:
   إذا سجل الدخول مستخدم لا يملك أدوار نظام عامة (مثل حساب `guest` أو مستخدم جديد عادي لم يُسند له دور مدير النظام `system-admin` في جدول `assigned_roles`)، فإن فحص `RequireGlobal` لـ `read-user-basic` يرفض وصوله ويرد بـ 403.

2. **عدم إدراج/بذر البيانات المطلوبة في قاعدة البيانات (Missing Seed Data)**:
   إذا كان المستخدم هو `admin` ولكن لم تتضمن قاعدة البيانات الصفوف المبذورة في `assigned_roles` (مثل البيانات المعرفة في الهجرة `000007_identity_bootstrap_and_projects_seed.up.sql`), فإن جدول `assigned_roles` يكون فارغاً ويُرفض طلب `admin` أيضاً بـ 403.

3. **اشتكاك منطقي بين حماية المسار ووظيفته (Route Authorization Design)**:
   دالة المعالج `handleUserPermissions` مخصصة لإرجاع أدوار وصلاحيات **المستخدم الحالي نفسه** لتحديد الميزات المتاحة له في الواجهة (UI Gating). إخضاع مسار استعلام الشخص عن صلاحياته الخاصة لحاجز حماية يتطلب صلاحية عامة (`read-user-basic`) يتسبب بظهور 403 Forbidden للمستخدمين العاديين، بدلاً من إرجاع قائمة صلاحيات فارغة `200 OK` بحمولة مثل:

   ```json
   {
     "roleAssignments": [],
     "ownedProjectIds": []
   }
   ```

---

## 5. تحليل كتالوج الصلاحيات وتدقيق الصلاحية المناسبة (Permissions Catalog Audit)

### أ. فحص كتالوج الصلاحيات المتاحة (`internal/services/permissions/permissions.go`)

تم إجراء مسح شامل لكافة الصلاحيات المعرّفة في كتالوج النظام تحت نطاق المستخدمين (`UserEntity`):

| معرف الصلاحية (ID) | اسم الصلاحية (Name) | الوصف | IsGlobal |
| :--- | :--- | :--- | :--- |
| `read-user-basic` | Read User Basic | عرض المعلومات الأساسية للمستخدمين | `true` |
| `read-user-details` | Read User Details | عرض التفاصيل الكاملة لحسابات المستخدمين | `true` |
| `update-self` | Update Self | تعديل الملف الشخصي للمستخدم نفسه | `true` |
| `update-user` | Update User | تعديل حسابات المستخدمين | `true` |
| `create-user` | Create User | إنشاء حسابات مستخدمين جديدة | `true` |
| `delete-user` | Delete User | حذف حسابات المستخدمين | `true` |

> **النتيجة**: لا توجد صلاحية مخصصة في الكتالوج تحت اسم "قراءة الصلاحيات الذاتية" (`read-self-permissions`).

### ب. هل تحديد `read-user-basic` للمسار خاطئ منطقياً؟

**نعم، تحديد صلاحية `read-user-basic` لهذا المسار خاطئ منطقياً ومعمارياً:**

1. **طبيعة الصلاحية مقابل طبيعة المسار**:
   * `read-user-basic` هي **صلاحية إدارية/عامة (Global Admin Permission)** تُمنح للمستخدم لتمكينه من البحث وقراءة بيانات المستخدمين الآخرين بالنظام.
   * المسار `/api/users/me/permissions` هو **مسار خدمة ذاتية (Self-Service Identity Endpoint)** يُفترض أن يتاح لكل مستخدم قام بتسجيل الدخول ليستعلم عن صلاحياته الشخصية التي يملكها لتهيئة الواجهة (UI Gating).

2. **التناقض المنطقي (Logical Paradox)**:
   إذا كان على المستخدم أن يملك صلاحية عامة إدارية (`read-user-basic`) فقط لمجرد أن يسأل الخادم *"ما هي الصلاحيات التي أملكها؟"*:
   * سيمنع الخادم أي مستخدم عادي أو ضيف من الوصول للمسار ويعيد `403 Forbidden`.
   * لن تتمكن الواجهة الأمامية (Frontend) من معرفة صلاحيات هذا المستخدم في المشاريع المختلفة، بينما الصواب أن يُرجع الخادم `200 OK` بقائمة أدواره المتاحة أو قائمة فارغة `{"roleAssignments": [], "ownedProjectIds": []}`.

---

## 6. الحل المعماري الموصى به (Recommended Architectural Solution)

بما أن المعالج `handleUserPermissions` يستخرج معرّف المستخدم الموثّق تلقائياً من سياق الجلسة (`GetCurrentUserID(request.Context())`)، فإن المسار يحتاج فقط إلى المصادقة (`protector`) ولا يحتاج إلى فحص صلاحية وسيطة عامة (`RequireGlobal`).

### التعديل البرمجي الموصى به

في الملف [`internal/httphandlers/handles/users/init.go`](file:///C:/Users/HC/Documents/projects/youtrack_system/youtrack_backend/internal/httphandlers/handles/users/init.go):

```diff
 router := parent.Route("/api/users", func(api chi.Router) {
     api.Use(protector)
     api.Get("/me", permissionRequired(perms.PermUserReadUserDetails, handler.handleMe))
-    api.Get("/me/permissions", permissionRequired(perms.PermUserReadUserBasic, handler.handleUserPermissions))
+    api.Get("/me/permissions", handler.handleUserPermissions)
 })
```

### فوائد هذا التعديل

1. يظل المسار محميًا بالمصادقة العامة للراوتر `api.Use(protector)`.
2. يستطيع أي مستخدم قام بتسجيل الدخول بنجاح استعلام صلاحياته الخاصة بـ `200 OK`.
3. المستخدمون الذين ليس لديهم أدوار مُسندة سيحصلون على `200 OK` بقائمة فارغة، مما يسمح للواجهة بالعمل بشكل طبيعي ودون أخطاء 403 Forbidden.
