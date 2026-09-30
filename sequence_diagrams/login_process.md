# مخطط تسلسل عملية تسجيل الدخول (Login Process)

هذا المخطط يوضح التفاعل بين مكونات النظام المختلفة عند محاولة مستخدم تسجيل الدخول.

```mermaid
sequenceDiagram
    autonumber
    
    actor User as "المستخدم"
    participant UI as "واجهة التطبيق"
    participant API as "AuthHandler (API)"
    participant App as "YouTrackApp (Logic)"
    participant DB as "UserStore (DB)"
    participant JWT as "JWT Service"

    User->>UI: إدخال اسم المستخدم وكلمة المرور
    UI->>+API: POST /api/auth/login {login, password}
    
    Note right of API: فك تشفير الطلب (JSON Decode)
    
    alt بيانات غير صالحة (فشل Decode)
        API-->>UI: 400 Bad Request
        UI-->>User: خطأ في صياغة الطلب
    else بيانات صالحة
        API->>+App: Authenticate(ctx, login, password)
        
        App->>+DB: GetByLogin(ctx, login)
        DB-->>-App: بيانات المستخدم (مع PasswordHash)
        
        alt المستخدم غير موجود
            App-->>API: 401 Unauthorized (invalid credentials)
        else المستخدم موجود
            App->>App: bcrypt.CompareHashAndPassword(hash, password)
            
            alt كلمة المرور خاطئة
                App-->>API: 401 Unauthorized (invalid credentials)
            else كلمة المرور صحيحة
                App-->>-API: User Object (بدون Hash)
                
                API->>+JWT: signToken(user)
                Note over JWT: توليد Claims (sub, login, roles, exp)
                JWT-->>-API: Signed JWT Token
                
                API-->>-UI: 200 OK {token, user, expires_in}
                UI-->>User: نجاح تسجيل الدخول والتوجيه
            end
        end
    end
```

## ملاحظات فنية

- يتم استخدام **bcrypt** لمقارنة كلمة المرور بأمان.
- يتم إصدار توكن **JWT** بصلاحية 72 ساعة.
- يتم تجريد `PasswordHash` من كائن المستخدم قبل إرساله في الاستجابة لضمان الخصوصية.
