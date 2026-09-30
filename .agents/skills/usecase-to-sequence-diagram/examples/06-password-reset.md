# مثال 6: إعادة تعيين كلمة المرور (Password Reset)

## وصف الـ Use Case

**العنوان**: إعادة تعيين كلمة مرور منسية  
**الممثل الرئيسي**: المستخدم  
**المسار الرئيسي**:
1. المستخدم يضغط "نسيت كلمة المرور"
2. إدخال البريد الإلكتروني
3. النظام يولّد رابط إعادة التعيين (مع Token ومدة صلاحية)
4. إرسال الرابط بالبريد
5. المستخدم يفتح الرابط
6. التحقق من صلاحية التوكن
7. إدخال كلمة المرور الجديدة
8. تحديث كلمة المرور وإبطال التوكن

---

## المخطط

```mermaid
sequenceDiagram
    autonumber
    actor User as "المستخدم"
    participant Browser as "المتصفح"
    participant API as "خادم API"
    participant DB as "قاعدة البيانات"
    participant Email as "خدمة البريد"

    %% ======= طلب إعادة التعيين =======
    User->>Browser: الضغط على "نسيت كلمة المرور"
    Browser-->>User: عرض نموذج إدخال البريد
    User->>Browser: إدخال البريد الإلكتروني
    Browser->>API: POST /forgot-password {email}
    activate API

    API->>DB: البحث عن المستخدم بالبريد
    activate DB
    DB-->>API: بيانات المستخدم
    deactivate DB

    note over API: يتم الرد بنفس الرسالة سواء<br/>وُجد المستخدم أم لا (أمان)

    opt المستخدم موجود
        API->>API: توليد resetToken + تحديد صلاحية (30 دقيقة)
        API->>DB: حفظ resetToken مع expiry
        activate DB
        DB-->>API: تم الحفظ
        deactivate DB
        API->>Email: إرسال رابط إعادة التعيين
        activate Email
        Email--)User: بريد إلكتروني مع الرابط
        deactivate Email
    end

    API-->>Browser: "تم إرسال رابط إعادة التعيين إذا كان البريد مسجلاً"
    deactivate API

    %% ======= إعادة التعيين =======
    User->>Browser: فتح رابط إعادة التعيين من البريد
    Browser->>API: GET /reset-password?token=abc123
    activate API
    API->>DB: التحقق من صلاحية التوكن
    activate DB
    DB-->>API: حالة التوكن
    deactivate DB

    alt التوكن صالح وغير منتهي
        API-->>Browser: عرض نموذج كلمة المرور الجديدة
        User->>Browser: إدخال كلمة المرور الجديدة (مرتين)
        Browser->>API: POST /reset-password {token, newPassword}

        API->>API: التحقق من قوة كلمة المرور
        API->>DB: تحديث كلمة المرور + إبطال التوكن
        activate DB
        DB-->>API: تم التحديث
        deactivate DB

        API-->>Browser: 200 - تم تغيير كلمة المرور بنجاح
        Browser-->>User: توجيه لصفحة تسجيل الدخول ✅

    else التوكن منتهي أو غير صالح
        API-->>Browser: 400 - الرابط منتهي الصلاحية
        Browser-->>User: "الرابط منتهي، اطلب رابطاً جديداً" ❌
    end
    deactivate API
```

---

## النقاط التعليمية

1. **`opt`**: العملية اختيارية — يتم إرسال البريد فقط إذا وُجد المستخدم
2. **ملاحظة أمنية**: الرد بنفس الرسالة لمنع تخمين البريد (Enumeration Attack)
3. **`note over` مع سطور متعددة**: استخدام `<br/>` لكسر السطر
4. **رسالة غير متزامنة `--)` **: إرسال البريد fire-and-forget
5. **نمط Token + Expiry**: نمط شائع في الأمان
