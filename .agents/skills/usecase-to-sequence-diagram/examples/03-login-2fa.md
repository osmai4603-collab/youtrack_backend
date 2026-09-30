# مثال 3: تسجيل دخول مع مصادقة ثنائية (2FA)

## وصف الـ Use Case

**العنوان**: تسجيل دخول آمن مع رمز OTP  
**الممثل الرئيسي**: المستخدم  
**الشروط المسبقة**: المستخدم مسجل ولديه رقم هاتف مفعّل  
**المسار الرئيسي**:

1. المستخدم يدخل البريد وكلمة المرور
2. المتصفح يرسل البيانات لخادم API
3. API يبحث عن المستخدم في قاعدة البيانات
4. التحقق من كلمة المرور
5. توليد رمز OTP وإرساله عبر SMS
6. المستخدم يدخل رمز OTP
7. التحقق من OTP وإنشاء JWT Token
8. توجيه المستخدم للوحة التحكم

**المسارات البديلة**:

- كلمة المرور خاطئة → رفض فوري
- رمز OTP خاطئ → رفض

---

## المخطط

```mermaid
sequenceDiagram
    autonumber
    actor User as "المستخدم"
    participant Browser as "المتصفح"
    participant API as "خادم API"
    participant Auth as "خدمة المصادقة"
    participant SMS as "خدمة الرسائل"
    participant DB as "قاعدة البيانات"

    %% ======= المرحلة 1: إدخال بيانات الدخول =======
    User->>Browser: إدخال البريد وكلمة المرور
    activate Browser
    Browser->>API: POST /login {email, password}
    activate API

    %% ======= المرحلة 2: التحقق من المستخدم =======
    API->>DB: البحث عن المستخدم بالبريد
    activate DB
    DB-->>API: بيانات المستخدم (أو لا يوجد)
    deactivate DB

    API->>Auth: التحقق من كلمة المرور (hash comparison)
    activate Auth
    Auth-->>API: نتيجة التحقق
    deactivate Auth

    alt كلمة المرور صحيحة
        %% ======= المرحلة 3: إرسال OTP =======
        API->>Auth: توليد رمز OTP
        activate Auth
        Auth->>SMS: إرسال OTP إلى رقم الهاتف
        activate SMS
        SMS-->>Auth: تم الإرسال
        deactivate SMS
        Auth-->>API: في انتظار OTP
        deactivate Auth

        API-->>Browser: 200 - طلب إدخال رمز OTP
        Browser-->>User: عرض شاشة إدخال OTP

        %% ======= المرحلة 4: التحقق من OTP =======
        User->>Browser: إدخال رمز OTP
        Browser->>API: POST /verify-otp {otp_code}
        API->>Auth: التحقق من OTP
        activate Auth

        alt OTP صحيح
            Auth-->>API: تحقق ناجح
            deactivate Auth
            API->>Auth: إنشاء JWT Token
            activate Auth
            Auth-->>API: JWT Token
            deactivate Auth
            API-->>Browser: 200 OK + Set-Cookie: token
            Browser-->>User: توجيه للوحة التحكم ✅
        else OTP خاطئ
            Auth-->>API: OTP غير صحيح
            deactivate Auth
            API-->>Browser: 401 Unauthorized
            Browser-->>User: رسالة "رمز OTP غير صحيح" ❌
        end

    else كلمة المرور خاطئة
        API-->>Browser: 401 Unauthorized
        Browser-->>User: "بيانات الدخول غير صحيحة" ❌
    end

    deactivate API
    deactivate Browser
```

---

## النقاط التعليمية

1. **مشاركون متعددون**: 6 مشاركين يمثلون الطبقات المختلفة (UI, API, Auth, SMS, DB)
2. **تسمية الرسائل بأسلوب API**: `POST /login {email, password}` بدلاً من وصف عام
3. **`alt` متداخل**: المستوى الأول لكلمة المرور، الثاني لـ OTP
4. **رموز إيموجي**: ✅ و ❌ لتوضيح النجاح والفشل بصرياً
5. **ترتيب المشاركين**: User → Browser → API → Auth → SMS → DB (من الواجهة للخلفية)
