# مثال 1: نظام إدارة المستشفى

## وصف الـ Use Case

**العنوان**: حجز موعد وإجراء فحص طبي  
**الممثل الرئيسي**: المريض  
**الشروط المسبقة**: المريض مسجل في النظام  
**المسار الرئيسي**:
1. المريض يطلب حجز موعد من الاستقبال
2. الاستقبال يتحقق من المواعيد المتاحة
3. تأكيد الموعد وحفظ الحجز
4. المريض يحضر للفحص
5. الطبيب يسترجع السجل الطبي
6. الطبيب يجري الفحص ويسجل التشخيص
7. إذا لزم الأمر، يكتب وصفة طبية
8. الصيدلية تسلم الأدوية

**المسار البديل**:
- لا توجد مواعيد متاحة → اقتراح بديل

---

## تحديد المشاركين

| المشارك | النوع | الدور |
|---------|-------|-------|
| المريض | Actor | يبدأ التفاعل |
| الاستقبال | Participant | يدير المواعيد |
| الطبيب | Participant | يجري الفحص |
| الصيدلية | Participant | تصرف الأدوية |
| قاعدة البيانات | Participant | تخزن البيانات |

---

## المخطط

```mermaid
sequenceDiagram
    autonumber
    actor Patient as "المريض"
    participant Reception as "الاستقبال"
    participant Doctor as "الطبيب"
    participant Pharmacy as "الصيدلية"
    participant DB as "قاعدة البيانات"

    %% ======= مرحلة الحجز =======
    Patient->>Reception: طلب حجز موعد
    activate Reception
    Reception->>DB: التحقق من المواعيد المتاحة
    activate DB
    DB-->>Reception: قائمة المواعيد
    deactivate DB

    alt مواعيد متاحة
        Reception-->>Patient: تأكيد الموعد
        Reception->>DB: حفظ الحجز
        activate DB
        DB-->>Reception: تم الحفظ
        deactivate DB
    else لا مواعيد متاحة
        Reception-->>Patient: لا مواعيد متاحة، اقتراح بديل
    end
    deactivate Reception

    %% ======= مرحلة الفحص =======
    Patient->>Doctor: الحضور للفحص
    activate Doctor
    Doctor->>DB: استرجاع السجل الطبي
    activate DB
    DB-->>Doctor: السجل الطبي
    deactivate DB
    Doctor->>Doctor: إجراء الفحص والتشخيص

    note over Doctor: تسجيل نتائج الفحص
    Doctor->>DB: تسجيل التشخيص
    activate DB
    DB-->>Doctor: تم التسجيل
    deactivate DB

    %% ======= مرحلة الوصفة (اختيارية) =======
    opt يحتاج وصفة طبية
        Doctor->>Pharmacy: إرسال الوصفة
        activate Pharmacy
        Pharmacy->>DB: تسجيل الوصفة
        activate DB
        DB-->>Pharmacy: تم التسجيل
        deactivate DB
        Pharmacy-->>Patient: تسليم الأدوية
        deactivate Pharmacy
    end

    Doctor-->>Patient: نصائح وتعليمات
    deactivate Doctor
```

---

## النقاط التعليمية

1. **استخدام `alt/else`**: لتمثيل حالة توفر المواعيد أو عدمها
2. **استخدام `opt`**: لتمثيل خطوة اختيارية (الوصفة الطبية)
3. **استخدام `note over`**: لإضافة توضيح فوق مشارك
4. **الرسائل الذاتية**: `Doctor->>Doctor` تمثل عملية داخلية (الفحص)
5. **التعليقات**: `%%` لتقسيم المخطط إلى مراحل واضحة
