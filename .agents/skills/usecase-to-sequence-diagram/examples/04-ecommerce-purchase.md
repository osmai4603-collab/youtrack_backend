# مثال 4: طلب شراء من متجر إلكتروني

## وصف الـ Use Case

**العنوان**: عملية شراء كاملة من متجر إلكتروني  
**الممثل الرئيسي**: العميل  
**الشروط المسبقة**: العميل مسجل الدخول  
**المسار الرئيسي**:
1. العميل يتصفح المنتجات
2. إضافة منتج للسلة (مع فحص المخزون)
3. متابعة للدفع وعرض ملخص السلة
4. إدخال بيانات الدفع
5. معالجة الدفع
6. إنشاء الطلب وتحديث المخزون (بالتوازي)
7. إرسال إيصال بالبريد الإلكتروني

**المسار البديل**:
- فشل الدفع → عرض رسالة خطأ

---

## المخطط

```mermaid
sequenceDiagram
    autonumber
    actor Customer as "العميل"
    participant Web as "المتجر الإلكتروني"
    participant Cart as "خدمة السلة"
    participant Inventory as "خدمة المخزون"
    participant Payment as "بوابة الدفع"
    participant Order as "خدمة الطلبات"
    participant Email as "خدمة البريد"

    %% ======= مرحلة التصفح والإضافة =======
    Customer->>Web: تصفح المنتجات
    activate Web
    Customer->>Web: إضافة منتج للسلة
    Web->>Cart: addItem(productId, qty)
    activate Cart
    Cart->>Inventory: checkStock(productId)
    activate Inventory
    Inventory-->>Cart: متوفر (الكمية المتاحة)
    deactivate Inventory
    Cart-->>Web: تمت الإضافة
    deactivate Cart
    Web-->>Customer: تحديث أيقونة السلة

    %% ======= مرحلة الدفع =======
    Customer->>Web: الانتقال لصفحة الدفع
    Web->>Cart: getCartSummary()
    activate Cart
    Cart-->>Web: ملخص السلة والمبلغ الإجمالي
    deactivate Cart
    Web-->>Customer: عرض صفحة الدفع

    Customer->>Web: إدخال بيانات الدفع وتأكيد
    Web->>Payment: processPayment(amount, cardInfo)
    activate Payment

    alt الدفع ناجح
        Payment-->>Web: تأكيد الدفع (transactionId)
        deactivate Payment

        note over Web: إنشاء الطلب وحجز المخزون بالتوازي

        %% ======= عمليات متوازية =======
        par إنشاء الطلب
            Web->>Order: createOrder(cartItems, transactionId)
            activate Order
            Order-->>Web: تأكيد الطلب (orderId)
            deactivate Order
        and تحديث المخزون
            Web->>Inventory: reserveStock(items)
            activate Inventory
            Inventory-->>Web: تم الحجز
            deactivate Inventory
        end

        %% ======= إرسال الإيصال =======
        Web->>Email: sendOrderConfirmation(orderId, email)
        activate Email
        Email--)Customer: إيصال بالبريد الإلكتروني
        deactivate Email

        Web-->>Customer: عرض صفحة "تم الطلب بنجاح" ✅

    else فشل الدفع
        Payment-->>Web: رفض الدفع (سبب الرفض)
        deactivate Payment
        Web-->>Customer: عرض رسالة خطأ الدفع ❌
    end

    deactivate Web
```

---

## النقاط التعليمية

1. **`par/and`**: تنفيذ متوازي لإنشاء الطلب وتحديث المخزون في نفس الوقت
2. **رسالة غير متزامنة `--)` **: البريد الإلكتروني يُرسل بدون انتظار تأكيد
3. **أسماء دوال**: `addItem()`, `getCartSummary()`, `processPayment()` بأسلوب برمجي
4. **`note over`**: ملاحظة توضيحية قبل العمليات المتوازية
5. **7 مشاركين**: يوضح كيفية التعامل مع بنية Microservices
