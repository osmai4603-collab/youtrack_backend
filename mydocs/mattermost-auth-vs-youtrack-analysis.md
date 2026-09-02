# تحليل الاختلافات بين auth و authorization في Mattermost ونسخة المشروع الحالي

## 1) ما الذي يوجد في Mattermost

في مشروع Mattermost، ملفات الأمان داخل مجلد app واضحة ومقسّمة بشكل منتظم:

- [mattermost/server/channels/app/authentication.go](../mattermost/server/channels/app/authentication.go)
- [mattermost/server/channels/app/authorization.go](../mattermost/server/channels/app/authorization.go)

### أ) authentication

تتضمن هذه الطبقة:

- التحقق من صحة كلمة المرور
- فحص أمان كلمات المرور
- دعم hashers متعددة
- تحديث كلمة المرور عند الحاجة
- فحص MFA
- حساب محاولات تسجيل الدخول
- التحكم في عدد محاولات الفشل
- دعم LDAP و SSO و sessions

في Mattermost، تنفيذ كلمة المرور لا يقتصر على مقارنة نصية بسيطة، بل يتضمن:

- `users.IsPasswordValidWithSettings`
- `hashers.GetHasherFromPHCString`
- `CompareHashAndPassword`
- `migratePassword` لتحديث خوارزمية التجزئة عند الحاجة
- `TryIncrementFailedPasswordAttempts` و `UpdateFailedPasswordAttempts`
- `CheckUserMfa` و `CheckUserPostflightAuthenticationCriteria`

وهذا يدل على أن Mattermost يعامل تسجيل الدخول كـ نظام أمان كامل، وليس مجرد تحقق بسيط من اسم المستخدم وكلمة المرور.

### ب) authorization

في Mattermost، إذن الوصول ليس مجرد هل المستخدم مسجّل؟ بل:

- هل له صلاحية معينة؟
- هل لديه دور معين؟
- هل هو عضو في فريق؟
- هل هو عضو في قناة؟
- هل هو admin؟
- هل هو مستخدم مُقيّد؟

الملف يضم وظائف مثل:

- `SessionHasPermissionTo`
- `SessionHasPermissionToAny`
- `SessionHasPermissionToTeam`
- `SessionHasPermissionToChannel`
- `SessionHasPermissionToChannels`
- `SessionHasPermissionToGroup`
- `SessionHasPermissionToChannelByPost`

وهذا يعني أن Mattermost يستخدم RBAC (Role-Based Access Control) مع تجميع مستوى:

1. نظام كامل
2. فريق
3. قناة
4. مجموعة
5. مستخدم

## 2) ما الذي يوجد في مشروعنا

ملف المصادقة في مشروعنا محدود جدًا:

- [channels/app/authentication.go](../channels/app/authentication.go)

يوجد فقط:

- `Register`
- `Authenticate`

والمنطق فيه بسيط جدًا:

- التحقق من عدم خلو اسم المستخدم وكلمة المرور
- جلب المستخدم بواسطة `GetByLogin`
- مقارنة الباسورد باستخدام bcrypt
- تكوين بعض الحقول الأساسية

لا يوجد:

- فحص معايير كلمة المرور
- MFA
- lockout بعد محاولات فشل
- sessions
- access tokens
- validation لتوكنات الدخول
- مسارات متعددة للتسجيل/مصادقة LDAP أو OAuth

## 3) الاختلافات الأساسية

### 1. Mattermost: نظام auth كامل
Mattermost يوفر:

- تكوين كلمات مرور وفق سياسة المؤسسة
- دعم التحديث الآمن لخوارزمية التجزئة
- محاولات فشل متدرجة
- MFA
- قيود على تسجيل الدخول
- تحقق من صلاحية الوصول حسب الدور
- عبر جلسات/أذونات/مجموعة/قناة

### 2. مشروعنا: auth أساسي جدًا
مشروعنا يحقق فقط نقطة الدخول الأساسية:

- إنشاء حساب
- تسجيل دخول

لكن لا توجد طبقات حماية أو سياسة أمان كافية.

## 4) النواقص في المشروع الحالي

### أ) غياب نظام الأذونات الحقيقي
في Mattermost توجد صلاحيات محددة لكل مستخدم، أما في مشروعنا:

- لا توجد ملفات authorization حقيقية
- لا توجد أدوار أو صلاحيات مركزية
- لا يوجد نظام `Permission` مشابه
- لا توجد فحوصات على مستوى الفريق/القناة/المنظمة

### ب) غياب إدارة الجلسات
في Mattermost توجد جلسات قوية ومعالجة مستمرة لها. في مشروعنا:

- لا يوجد `Session` model واضح
- لا يوجد JWT validation كامل
- لا توجد إدارة لتوقيت الجلسات
- لا يوجد logout أو revoke session

### ج) غياب MFA و lockout
في Mattermost يوجد:

- فحص MFA
- حماية من brute-force
- تسجيل عدد محاولات الفشل

في مشروعنا لا يوجد ذلك.

### د) غياب سياسات كلمة المرور
Mattermost يتحقق من:

- طول كلمة المرور
- التعقيد
- نوع التجزئة
- إمكانية التحديث التلقائي

في مشروعنا يتم استخدام bcrypt فقط، دون سياسات أو محو/تحديث مناسب.

### هـ) غياب معاملة الحقول الحساسة بشكل آمن
في Mattermost يتم التعامل مع كلمة المرور كـ بيانات حساسة جدًا، مع فحص نوعها، معالجتها، وتحديثها عند الحاجة. في مشروعنا:

- لا يوجد `checkUserPassword`
- لا يوجد `migratePassword`
- لا يوجد حماية من hash غير متوافق
- لا يوجد نمط استبدال/تجديد آمن

### و) غياب بنية permission hierarchy
Mattermost يمرر من `Session` إلى `Roles` إلى `Permission` إلى `Channel/Team/Group`.

في مشروعنا لا توجد هذه البنية، وبالتالي لا توجد قاعدة للسيطرة على الوصول إلى الموارد.

## 5) الخلاصة

إذا قارنا الأنظمة فإن Mattermost يطبق نظام أمان متكامل، بينما مشروعنا يطبق فقط “مصادقة أولية” في أبسط صورة.

المقارنة المختصرة:

| العنصر | Mattermost | المشروع الحالي |
|---|---|---|
| تسجيل دخول | متقدم | أساسي |
| كلمة المرور | سياسة + hashers + migration | bcrypt فقط |
| MFA | موجود | غير موجود |
| Lockout | موجود | غير موجود |
| Sessions | موجودة ومتكاملة | غير موجودة/بسيطة |
| Authorization | موجود جدًا | شبه غائب |
| Permissions per role | موجود | غير موجود |
| Team/Channel access control | موجود | غير موجود |

## 6) النتيجة العملية

الهيكل الحالي يشبه Mattermost في النمط العام فقط، لكنه لا يطبق طبقة الأمان نفسها. فالمشروع يملك غلاف تقني يشبه Mattermost، لكنه ما زال في مرحلة نمذجة أولية للـ auth، وليس نظام auth/authorization جاهز للإنتاج.

إذا كان الهدف هو بناء نظام مشابه لـ Mattermost بشكل حقيقي، فينبغي إضافة:

1. نظام sessions
2. system permissions
3. roles/permissions per team/channel
4. MFA
5. brute-force protection
6. password policy
7. auth middleware
8. authorization checks on API routes

## 7) ملاحظة مهمة

حقيقة أن ملفات `authentication` و `authorization` موجودة في Mattermost لا تعني أن كل مشروع يجب أن يكررها حرفيًا. لكن إذا كان المشروع يطمح إلى نسخ بنية Mattermost، فالفجوة الحالية واضحة: التطبيق الحالي يتعامل مع تسجيل الدخول كوظيفة CRUD بسيطة، وليس كطبقة أمن كاملة.
