# خطة تنفيذ دورة حياة الخادم

## 1. الهدف

تحويل دورة حياة YouTrack Backend من دوال متفرقة إلى مسار تشغيل موحد يملك فيه الخادم HTTP listener والمهام الخلفية والاعتماديات، ويعلن readiness فقط بعد نجاح startup، ويغلق الموارد بترتيب قابل للاختبار.

المرجع المقارن هو Mattermost Server، لكن الخطة لا تهدف إلى نسخ ميزات Mattermost غير المطلوبة مثل WebSocket أو cache أو search أو metrics.

## 2. قواعد التنفيذ

- لا readiness قبل نجاح bind وبدء المكونات الحرجة.
- كل resource يبدأه الخادم يجب أن يملك stop/close وانتظارا واضحا.
- signal handling يملكه entry point مرة واحدة فقط.
- لا تغلق dependency قبل workers التي تستخدمها.
- لا يتوقف shutdown عند أول خطأ؛ يجب إكمال التنظيف وتجميع الأخطاء.
- كل اختبار lifecycle يجب أن يحاكي ترتيب التشغيل الفعلي في production.

## 3. المراحل المتسلسلة

### المرحلة 0: تثبيت العقد

1. اعتماد حالات lifecycle: `Created`, `Validating`, `Starting`, `Running`, `Stopping`, `Stopped`, و`Failed`.
2. تحديد مالك كل مورد: HTTP server/listener، local server، channels tasks، platform/store، workers، وroot context.
3. تثبيت invariants: لا readiness قبل bind، لا goroutine بلا cancellation وwait، ولا resource بلا Close أو Stop.
4. إبقاء WebSocket وcache وsearch وmetrics خارج المسار الأساسي حتى يثبت أنها متطلبات منتج.

**مخرج المرحلة:** عقد lifecycle مكتوب وقائمة resources مع مالكها وسلوك إيقافها.

### المرحلة 1: توحيد startup وحقن الاعتماديات

الملفات الرئيسية:

- `channels/app/youtrack_server.go`
- `channels/app/options.go`
- `cmd/server/main.go`
- `channels/api/router.go`

التنفيذ:

1. عدم تجاهل خطأ `platform.New()`.
2. تركيب API router قبل بدء serving وقبل إعلان readiness.
3. جعل `YouTrackServer` يملك إعداد HTTP server وlistener.
4. فصل وصف الاعتماديات في options عن acquisition قدر الإمكان.
5. إضافة listener factory أو injection لاختبار فشل bind.

**التحقق:** إنشاء server صحيح، التأكد من اكتمال handler قبل startup، وإثبات أن فشل platform أو listener يعيد خطأ ولا يعلن readiness.

### المرحلة 2: إصلاح `Run` وملكية الإشارات

الملفات الرئيسية:

- `channels/app/youtrack_server.go`
- `channels/app/youtrack_channels.go`
- `cmd/server/main.go`

التنفيذ:

1. جعل `Run(ctx)` يبدأ startup ثم serving ويستجيب لإلغاء context باستدعاء shutdown.
2. إزالة `signal.Notify` وقرار إيقاف العملية من `YouTrackChannels`.
3. إبقاء signal ownership في `cmd/server/main.go` عبر root context واحد.
4. اعتبار `http.ErrServerClosed` نتيجة طبيعية.
5. تنظيف الموارد المنشأة جزئيا عند فشل startup.

**التحقق:** تشغيل `Run` على listener عشوائي، إلغاء context، والتأكد من انتهاء serving دون deadlock.

### المرحلة 3: readiness وhealth

الملفات الرئيسية:

- `channels/api/health.go`
- `channels/api/router.go`
- `channels/app/youtrack_server.go`

التنفيذ:

1. ربط readiness بنجاح bind وبدء dependencies الحرجة.
2. إزالة readiness فور بدء shutdown.
3. فصل `/live` عن `/ready`، أو توثيق `/health` بوضوح إذا كان الحفاظ عليه ضروريا.
4. إصلاح local health حتى لا يعيد نجاحا دائما إذا كان يمثل readiness.
5. فحص store والاعتماديات الحرجة ضمن latency مناسبة.

**التحقق:** 503 قبل startup، 200 بعد bind، و503 أثناء وبعد shutdown؛ مع منع readiness عند فشل DB أو bind.

### المرحلة 4: إدارة workers والمهام المجدولة

الملفات الرئيسية:

- `channels/app/youtrack_channels.go`
- `channels/app/platform/platform_service.go`

التنفيذ:

1. إضافة worker registry باستخدام `sync.WaitGroup` أو abstraction مكافئ.
2. تسجيل كل goroutine عبر wrapper يحدد cancellation وسياسة الخطأ.
3. إضافة stop signal موحد للمهام التابعة لـ `YouTrackChannels`.
4. إلغاء `dndTask` و`schedulerPostTask` وجميع المهام المجدولة.
5. جعل `Stop` idempotent وآمنا قبل اكتمال `Start`.
6. انتظار workers قبل إغلاق store.
7. تحديد سياسة أخطاء workers: إيقاف الخادم أو التسجيل والاستمرار.

**التحقق:** مهمة طويلة تستجيب للإلغاء، وإثبات أن worker ينتهي قبل إغلاق store.

### المرحلة 5: graceful shutdown وتجميع الأخطاء

الترتيب المقترح:

```text
mark unready
  -> stop accepting traffic
  -> drain HTTP/local/WS connections
  -> cancel tasks
  -> stop workers and wait
  -> close platform/store
  -> flush logger
  -> return aggregated errors
```

التنفيذ:

1. إيقاف استقبال traffic قبل إيقاف services.
2. منع early return عند فشل مرحلة.
3. تجميع أخطاء الإغلاق وإرجاعها للمستدعي.
4. وضع timeout للـ drain والـ worker wait.
5. الحفاظ على idempotency في shutdown المتكرر والمتزامن.
6. إغلاق وإزالة local UNIX socket إن كان مستخدما.

**التحقق:** مكونات تفشل بشكل متعمد، مع التأكد من إغلاق بقية الموارد وعدم حدوث deadlock.

### المرحلة 6: الاختبارات والتوثيق

إضافة أو تحديث اختبارات:

- invalid configuration.
- failed bind.
- `Run(ctx)` cancellation.
- readiness timing.
- worker drain.
- scheduled task cancellation.
- concurrent shutdown.
- shutdown errors.
- local UNIX socket.
- production-shaped startup مطابق لـ `cmd/server/main.go`.

أوامر التحقق:

```bash
go test ./...
go test -race ./...
go vet ./...
go build ./...
```

وتوثيق قرارات lifecycle النهائية: timeouts، dependency criticality، worker error policy، ودلالة liveness/readiness.

### المرحلة 7: القدرات الاختيارية

لا تبدأ قبل نجاح المراحل السابقة.

1. تقييم WebSocket: router، auth handshake، connection registry، hub، shutdown.
2. تقييم cache وsearch وmetrics وjob server بحسب متطلبات المنتج.
3. لكل capability، إضافة resource owner وstop signal وwait/error policy واختبار lifecycle.

## 4. خارطة التنفيذ الحالية

تم تنفيذ الجزء الأول من المرحلة 1 والمرحلة 2 جزئيا:

- `Run` يفتح listener ويستخدم `Server.Serve`.
- إلغاء context يؤدي إلى graceful shutdown.
- readiness لا تعلن في `Start`، بل بعد bind داخل `Run`.
- `main` يركب router وHTTP server قبل `Run`.
- إزالة process-level signal handling من `YouTrackChannels`.
- إلغاء `dndTask` و`schedulerPostTask` عند إيقاف القنوات.
- إضافة اختبارات failed bind وcontext cancellation وreadiness timing.

تم إكمال المرحلة 3:

- إضافة `/live` لفحص liveness.
- إضافة `/ready` لفحص readiness المرتبط بنجاح listener binding.
- إبقاء `/health` كمسار توافق يعكس readiness.
- توحيد دلالات API router وlocal router والمسار الاحتياطي داخل `Validate`.
- إضافة اختبارات readiness وliveness بعد startup وبعد shutdown، واختبار local router بدون server.

تم تنفيذ الجزء الأساسي من المرحلة 4:

- إضافة worker registry إلى `PlatformService` عبر `Go` و`Context` و`WaitForWorkers`.
- إلغاء workers وانتظارها قبل إغلاق store عبر `ShutdownContext`.
- ربط `YouTrackServer.Shutdown(ctx)` بسياق إغلاق المنصة.
- جعل `ScheduledTask.Cancel` آمنا عند الاستدعاء المتكرر.
- إضافة اختبارات worker cancellation/drain واختبار idempotent task cancellation.

تم تنفيذ المرحلة التالية الخاصة بحالة lifecycle وتجميع أخطاء الإغلاق:

- إضافة `LifecycleState` وحالات `created`, `starting`, `running`, `stopping`, `stopped`, و`failed`.
- إضافة `LifecycleState()` لقراءة الحالة بأمان.
- منع بدء خادم تم إيقافه أو ما زال في طور الإيقاف.
- تسجيل حالات startup الفاشل والتشغيل بعد bind والإيقاف النهائي.
- تجميع أخطاء HTTP وchannels وplatform أثناء shutdown باستخدام `errors.Join` بدلا من التوقف عند أول خطأ.
- إضافة assertions لاختبار انتقالات lifecycle الأساسية.

تم إكمال lifecycle الخاص بالـ local supervision API:

- إضافة local listener وlocal HTTP server كموارد اختيارية يملكها `YouTrackServer`.
- تشغيل local server داخل `Run` عند ربطه عبر `SetLocalServer`.
- إيقاف local server وإغلاق listener وحذف UNIX socket أثناء `Shutdown`.
- إضافة اختبار تكاملي يرسل `/ready` عبر UNIX domain socket ويتحقق من تنظيف socket.

تم تنفيذ تحسينات shutdown المتبقية:

- إضافة timeout مستقل لكل مرحلة من مراحل HTTP وlocal server وplatform.
- تحرير سياق timeout بعد انتهاء كل مرحلة لمنع بقاء timers معلقة.
- إضافة اختبار timeout عند تعذر worker على الانتهاء.
- إضافة اختبار concurrent shutdown للتحقق من idempotency وعدم حدوث deadlock.

تم إكمال مسار تشغيل workers الفعلية:

- إضافة `PlatformService.GoContext` لتشغيل worker مع `context.Context` المملوك للمنصة.
- إضافة `YouTrackChannels.StartWorker` كواجهة app لتشغيل workers عبر المنصة، بدلا من `go` مباشر.
- workers التي تبدأ بهذا المسار تدخل تلقائيا في `WaitForWorkers` قبل إغلاق `store`.
- إضافة اختبارات تثبت وصول cancellation وانتهاء worker قبل عودة shutdown.
- لا توجد حاليا worker إنتاجية أخرى منشأة داخل `channels/app` تحتاج نقلها؛ goroutines الحالية الخاصة بـ HTTP وlocal serving جزء من lifecycle نفسه.

المتبقي من الخطة: استخدام `StartWorker` لأي worker مستقبلية جديدة، وإضافة أي capabilities اختيارية مثل WebSocket أو metrics أو cache حسب متطلبات المنتج.

## 5. الملفات المستهدفة

- `channels/app/youtrack_server.go`
- `channels/app/youtrack_channels.go`
- `channels/app/platform/platform_service.go`
- `channels/app/options.go`
- `cmd/server/main.go`
- `channels/api/router.go`
- `channels/api/health.go`
- `channels/app/server_lifecycle_test.go`
- `channels/api/middleware_test.go`

## 6. معايير القبول

يعتبر lifecycle الأساسي مكتملا عندما:

1. يفشل startup بوضوح عند config أو dependency أو bind غير صالح.
2. لا تصبح الخدمة ready قبل امتلاك listener صالح.
3. يتوقف `Run(ctx)` عند إلغاء context.
4. لا تسجل طبقة domain process-level signals.
5. تلغى كل المهام والـ workers وينتظر انتهاؤها.
6. يكمل shutdown تنظيف الموارد عند فشل مرحلة منفردة.
7. تمر اختبارات `go test ./...` و`go test -race ./...` و`go vet ./...` و`go build ./...`.
