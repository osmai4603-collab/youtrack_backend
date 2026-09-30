# مثال 5: نظام حجز رحلات طيران

## وصف الـ Use Case

**العنوان**: حجز رحلة طيران عبر تطبيق السفر  
**الممثل الرئيسي**: المسافر  
**الشروط المسبقة**: المسافر مسجل في التطبيق  
**المسار الرئيسي**:
1. المسافر يبحث عن رحلات (من، إلى، تاريخ)
2. عرض الرحلات المتاحة
3. اختيار رحلة واختيار المقعد
4. إدخال بيانات الركاب
5. الدفع
6. إصدار التذكرة
7. إرسال تأكيد + تذكرة إلكترونية

**المسارات البديلة**:
- لا رحلات متاحة → اقتراح تواريخ بديلة
- فشل الدفع → إلغاء الحجز المؤقت
- المقعد محجوز (Race Condition) → اختيار مقعد آخر

---

## المخطط

```mermaid
sequenceDiagram
    autonumber
    actor Traveler as "المسافر"
    participant App as "تطبيق السفر"
    participant FlightSvc as "خدمة الرحلات"
    participant SeatSvc as "خدمة المقاعد"
    participant PaymentGW as "بوابة الدفع"
    participant TicketSvc as "خدمة التذاكر"
    participant NotifySvc as "خدمة الإشعارات"

    %% ======= البحث عن رحلات =======
    Traveler->>App: بحث (من: الرياض، إلى: دبي، تاريخ: 2024-03-15)
    activate App
    App->>FlightSvc: searchFlights(origin, dest, date)
    activate FlightSvc
    FlightSvc-->>App: قائمة الرحلات المتاحة (3 رحلات)
    deactivate FlightSvc

    alt رحلات متاحة
        App-->>Traveler: عرض الرحلات مع الأسعار

        %% ======= اختيار رحلة ومقعد =======
        Traveler->>App: اختيار الرحلة FL-205
        App->>SeatSvc: getAvailableSeats(flightId)
        activate SeatSvc
        SeatSvc-->>App: خريطة المقاعد المتاحة
        deactivate SeatSvc
        App-->>Traveler: عرض خريطة المقاعد

        Traveler->>App: اختيار المقعد 14A
        App->>SeatSvc: holdSeat(flightId, seatId, sessionId)
        activate SeatSvc

        alt المقعد متاح
            SeatSvc-->>App: تم حجز المقعد مؤقتاً (5 دقائق)
            deactivate SeatSvc

            %% ======= بيانات الراكب والدفع =======
            App-->>Traveler: نموذج بيانات الراكب
            Traveler->>App: إدخال بيانات الراكب (الاسم، جواز السفر)

            Traveler->>App: تأكيد الدفع
            App->>PaymentGW: charge(amount, paymentMethod)
            activate PaymentGW

            alt الدفع ناجح
                PaymentGW-->>App: تم الدفع (transactionId)
                deactivate PaymentGW

                %% ======= إصدار التذكرة =======
                App->>SeatSvc: confirmSeat(flightId, seatId)
                activate SeatSvc
                SeatSvc-->>App: تم تأكيد المقعد
                deactivate SeatSvc

                App->>TicketSvc: issueTicket(flight, passenger, seat)
                activate TicketSvc
                TicketSvc-->>App: التذكرة (ticketNumber: TK-98765)
                deactivate TicketSvc

                par إشعار العميل
                    App->>NotifySvc: sendConfirmation(email, ticketPDF)
                    activate NotifySvc
                    NotifySvc--)Traveler: بريد إلكتروني + تذكرة PDF
                    deactivate NotifySvc
                and إشعار SMS
                    App->>NotifySvc: sendSMS(phone, confirmationCode)
                    activate NotifySvc
                    NotifySvc--)Traveler: رسالة SMS بكود التأكيد
                    deactivate NotifySvc
                end

                App-->>Traveler: عرض "تم الحجز بنجاح" ✅

            else فشل الدفع
                PaymentGW-->>App: رفض الدفع
                deactivate PaymentGW
                App->>SeatSvc: releaseSeat(flightId, seatId)
                activate SeatSvc
                SeatSvc-->>App: تم تحرير المقعد
                deactivate SeatSvc
                App-->>Traveler: فشل الدفع - تم تحرير المقعد ❌
            end

        else المقعد محجوز (Race Condition)
            SeatSvc-->>App: المقعد غير متاح
            deactivate SeatSvc
            App-->>Traveler: المقعد محجوز، اختر مقعداً آخر
        end

    else لا رحلات متاحة
        App-->>Traveler: لا رحلات متاحة، اقتراح تواريخ بديلة
    end

    deactivate App
```

---

## النقاط التعليمية

1. **سيناريو واقعي معقد**: 7 مشاركين مع مسارات بديلة متعددة
2. **`alt` ثلاثي المستويات**: رحلات → مقعد → دفع
3. **`par/and`**: إرسال بريد و SMS بالتوازي
4. **Rollback Pattern**: عند فشل الدفع، يتم تحرير المقعد المحجوز مؤقتاً
5. **Race Condition**: التعامل مع حجز متزامن لنفس المقعد
6. **Hold Pattern**: حجز مؤقت (5 دقائق) قبل التأكيد
