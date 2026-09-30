# تقرير تحليل ومقارنة دورة حياة الخادم

## 1. النطاق

يقارن هذا التقرير دورة حياة الخادم داخل طبقة `channels/app` في مشروع YouTrack Backend مع دورة الحياة المنفذة في Mattermost Server الموجود في:

`/home/osm/mattermost/server`

يركز التحليل على الخادم كعملية طويلة العمر:

- بناء الاعتماديات والتحقق منها.
- ترتيب startup.
- تركيب routes وmiddleware.
- فتح HTTP listeners.
- تشغيل workers والمهام الخلفية.
- readiness وhealth.
- الإشارات والإلغاء.
- graceful shutdown.
- انتشار الأخطاء والاختبارات.

لا يعني استخدام Mattermost كمرجع ضرورة نسخ قدراته الاختيارية، مثل البريد أو cache أو search أو WebSocket، ما لم تكن مطلوبة في YouTrack.

## 2. الخلاصة التنفيذية

يمتلك YouTrack الهيكل الاسمي الأساسي:

```text
NewServer -> Validate -> Start -> Run -> Shutdown
```

لكن المسؤوليات موزعة بين `cmd/server/main.go` و`YouTrackServer` و`YouTrackChannels` وطبقة API. لذلك فإن lifecycle الحالي يمثل عقداً جيداً للاختبار، لكنه ليس orchestration مركزياً مكتملاً لكل الموارد.

Mattermost يجعل `Server` مالكاً لمعظم موارد lifecycle. يبني الخدمات بترتيب اعتماد واضح، ثم يجعل `Start()` مسؤولاً عن تشغيل الخدمات، فتح listeners، وتشغيل HTTP server. أما YouTrack فيقسم تهيئة الخدمات بين `Start()` وتشغيل الشبكة بين `Run()` وentry point.

## 3. دورة YouTrack الحالية

### 3.1. البناء والتحقق

الملفات الأساسية:

- `channels/app/youtrack_server.go`
- `channels/app/options.go`
- `channels/app/youtrack_channels.go`
- `channels/app/platform/platform_service.go`
- `cmd/server/main.go`

التسلسل:

```text
config.Load
  -> NewServer
  -> WithStore
  -> WithConfig
  -> WithJWTSecret
  -> NewChannels
  -> Validate
  -> platform.Start
  -> Channels.Start
  -> net.Listen
  -> HTTP Serve
```

`NewServer` ينشئ routers ويطبق الخيارات ويهيئ المنصة والقنوات. `Validate` يتحقق من الإعدادات وJWT وstore، ثم يضمن وجود routers وcontext و`http.Server`.

### 3.2. Start وRun

`Start()` يقوم حالياً بـ:

1. تغيير الحالة إلى `starting`.
2. تنفيذ `Validate()`.
3. تشغيل `platform.Start()`.
4. تشغيل `Channels().Start()`.

أما `Run(ctx)` فيقوم بـ:

1. استدعاء `Start()`.
2. فتح TCP listener باستخدام `net.Listen`.
3. حفظ listener داخل الخادم.
4. تشغيل `http.Server.Serve` في goroutine.
5. تعيين readiness إلى `true` بعد نجاح bind.
6. انتظار فشل server أو إلغاء `ctx`.
7. تنفيذ `Shutdown` عند الإلغاء أو الفشل.

الفصل بين `Start` و`Run` صالح تصميمياً، لكنه يتطلب أن يستخدم كل entry point المسار الصحيح. نقطة التشغيل الحالية في `cmd/server/main.go` تبني API router و`http.Server` ثم تستدعي `Run`.

### 3.3. Channels وworkers

`YouTrackChannels.Start()` يهيئ قناة الإيقاف فقط. توجد آلية عامة في `PlatformService` لتشغيل workers عبر:

- `Go`.
- `GoContext`.
- `WaitForWorkers`.
- `StartWorker` من طبقة Channels.

لكن لا توجد حالياً مجموعة production jobs أو schedulers فعلية تبدأ تلقائياً ضمن `Channels.Start()`.

## 4. دورة Mattermost المناظرة

الملفات المرجعية الأساسية:

- `/home/osm/mattermost/server/channels/app/server.go`
- `/home/osm/mattermost/server/cmd/mattermost/commands/server.go`

### 4.1. NewServer

`NewServer` في Mattermost يبني dependency graph واسعاً بترتيب معلن، ومن أمثلته:

1. Root وlocal routers.
2. Platform وHTTP service.
3. Enterprise services.
4. Users وTeams وProperties services.
5. Builtin property groups وhooks.
6. Platform startup.
7. Channels وApp.
8. Caches وtemplates وtranslations.
9. Telemetry وemail وfeature flags.
10. Jobs وlisteners وmigrations حسب الإعدادات.

الفكرة المهمة ليست عدد الخدمات، بل أن كل خدمة تنشأ بعد توفر اعتمادياتها، وأن أخطاء الإنشاء الحرجة تعاد إلى المستدعي.

### 4.2. Start

`Server.Start()` في Mattermost ينسق التشغيل الفعلي، ومنه:

- تشغيل inter-cluster وshared-channel services.
- تشغيل Channels والـ plugins.
- تشغيل cluster communication.
- تنفيذ تهيئة startup المطلوبة في قاعدة البيانات.
- اختبار mail وfile storage.
- إعادة تحميل config.
- تركيب Sentry وCORS وrate limiting وTLS middleware.
- إنشاء TCP listener.
- تشغيل HTTP أو HTTPS server.
- تشغيل local Unix socket عند تفعيله.
- تشغيل workers وjobs وschedulers وفق الإعدادات.

بعد نجاح bind يمكن للـ command layer إعلان أن الخادم أصبح جاهزاً.

### 4.3. Shutdown

إيقاف Mattermost مرتب حسب الاعتماديات:

```text
إزالة listeners
  -> إيقاف الخدمات العنقودية
  -> إيقاف HTTP/local servers
  -> إيقاف push وsearch وaudit
  -> إيقاف config وmetrics وemail
  -> إيقاف jobs وschedulers
  -> إيقاف Channels
  -> إيقاف Platform وStore
  -> flush وإيقاف logging
```

القاعدة الأساسية: لا تغلق dependency قبل إيقاف المكونات التي تستخدمها.

## 5. جدول المقارنة

| المجال | YouTrack Backend | Mattermost Server |
|---|---|---|
| نقطة التشغيل | موزعة بين `Start` و`Run` وentry point | `Server.Start` هو حد التشغيل العملي |
| بناء الخدمات | Platform وChannels بالأساس | خدمات نطاق، jobs، caches، telemetry، cluster وغيرها |
| التحقق | `Validate` صريحة وتفحص config/store/JWT | تحقق موزع أثناء بناء وتشغيل الخدمات |
| HTTP listener | يفتح في `Run` | يفتح داخل `Start` |
| readiness | حالة صريحة `ready` | لا توجد state machine مماثلة داخل Server |
| local health | API موجود، لكن يحتاج ربط listener يدوياً | local mode جزء من startup عند تفعيله |
| workers | registry عامة، بلا jobs إنتاجية افتراضية | jobs وschedulers متعددة |
| الإشارات | `signal.NotifyContext` في command | signal handling في command بعد startup |
| TLS | غير موجود ضمن lifecycle الحالي | TLS وLet’s Encrypt وredirects |
| shutdown errors | تجميع عبر `errors.Join` | تسجيل الأخطاء غالباً مع إكمال الإغلاق |
| اختبار bind | موجود | bind failure يعاد من `Start` |

## 6. الفروقات والنواقص في YouTrack

### 6.1. Start لا يشغل دورة التشغيل كاملة

`Start()` يشغل Platform وChannels، لكنه لا يفتح listener ولا يبدأ HTTP serving. هذا يجعل callers بحاجة إلى معرفة الفرق بين `Start` و`Run`، ويزيد احتمال بناء مسار تشغيل غير مكتمل.

### 6.2. local supervision غير موصول افتراضياً

توجد الدوال:

- `api.NewLocalRouter`.
- `api.StartLocalAPI`.
- `YouTrackServer.SetLocalServer`.

لكن `cmd/server/main.go` لا يستدعي `StartLocalAPI` ولا `SetLocalServer`. لذلك فإن local `/live` و`/ready` لا يعملان في المسار الافتراضي إلا إذا ربطهما مستدعٍ آخر.

### 6.3. Channels.Start لا يبدأ workers فعلية

حقول المهام المجدولة موجودة، كما أن `Stop()` يلغي بعض المهام، لكن إنشاء وتشغيل هذه المهام ليس جزءاً من startup. النتيجة أن lifecycle يملك آلية الإيقاف أكثر مما يملك آلية التشغيل.

### 6.4. rollback عند فشل startup محدود

إذا نجح `platform.Start()` ثم فشل `Channels.Start()` أو مرحلة لاحقة، لا يوجد transaction واضح يعكس الموارد التي بدأت بالفعل. ينبغي تنفيذ cleanup جزئي منظم لكل مرحلة فاشلة.

### 6.5. حماية Start من التزامن غير مكتملة

يوجد `shutdownOnce`، لكن لا توجد حماية مكافئة على مستوى بدء الخادم. استدعاء `Start` أو `Run` بالتزامن قد يؤدي إلى تشغيل مكرر أو سباق على state وlisteners.

### 6.6. ملكية listener الرئيسي غير صريحة بما يكفي

يحفظ الخادم `listener`، لكن الإغلاق يعتمد أساساً على `http.Server.Shutdown`. يلزم تحديد ملكية listener ومسؤولية إغلاقه وتصفيره في كل مسار فشل أو إيقاف.

### 6.7. timeout الإيقاف يحتاج سياسة موحدة

يتم إنشاء timeout مستقل لكل مرحلة إيقاف. إذا كان parent context قريباً من الانتهاء فقد تفشل مراحل لاحقة فوراً. الأفضل تعريف timeout إجمالي أو ميزانية زمنية واضحة لدورة shutdown.

### 6.8. بعض أخطاء الإيقاف تضيع

في مسار فشل `Serve` يستدعي `Run` الإيقاف ثم يهمل خطأ `Shutdown`. هذا يضعف تشخيص الفشل، خصوصاً عندما يكون فشل الخدمة متبوعاً بفشل تنظيف مورد آخر.

### 6.9. constructors غير متسقة في معالجة الأخطاء

`NewServerWithOptions` يتجاهل خطأ `platform.New()`، بعكس `NewServer`. ينبغي توحيد عقد الإنشاء وعدم تجاهل الأخطاء المحتملة.

### 6.10. فحص readiness محدود

readiness تعتمد أساساً على نجاح bind ووجود flag. لا يوجد فحص runtime مستقل لخدمات حرجة بعد startup، مثل store أو أي external dependency مستقبلية.

## 7. الخطوات التنفيذية الأساسية لأي Server

### المرحلة 1: Configuration

- تحميل configuration.
- تطبيق defaults.
- التحقق من القيم المطلوبة.
- تحميل الأسرار بأمان.

### المرحلة 2: Dependency Construction

- إنشاء logger.
- إنشاء store والاتصال بقاعدة البيانات.
- تشغيل migrations عند الحاجة.
- إنشاء clients وcaches والخدمات.
- تسجيل owner لكل resource.

### المرحلة 3: Preflight Validation

- فحص صحة configuration.
- فحص database readiness.
- فحص external dependencies الحرجة.
- رفض startup قبل إنشاء listeners عند فشل dependency أساسية.

### المرحلة 4: Route and Middleware Assembly

- إنشاء root وlocal routers.
- تسجيل API وhealth routes.
- تركيب authentication وrecovery وlogging وCORS وrate limiting.
- تثبيت handler قبل فتح listener.

### المرحلة 5: Runtime Service Startup

- تشغيل platform.
- تشغيل domain services.
- تشغيل plugins وcluster services عند الحاجة.
- تشغيل workers وschedulers عبر registry قابلة للإلغاء والانتظار.

### المرحلة 6: Listener Binding

- فتح TCP أو TLS listener.
- فتح Unix socket عند الحاجة.
- إعادة خطأ bind فوراً إلى caller.
- تحديد ملكية كل listener بوضوح.

### المرحلة 7: Readiness

- `live`: العملية قادرة على الرد.
- `ready`: كل الاعتماديات الحرجة جاهزة، والـ listener مربوط، والخادم قادر على استقبال traffic.
- إزالة readiness فور بدء shutdown.

### المرحلة 8: Serving and Monitoring

- تشغيل servers في goroutines مراقبة.
- التقاط أخطاء serve.
- تحويل فشل server إلى shutdown منظم.
- منع workers غير المسجلة.

### المرحلة 9: Graceful Shutdown

الترتيب العام:

```text
mark unready
  -> stop accepting traffic
  -> drain HTTP/local/WS connections
  -> cancel workers and tasks
  -> wait for workers
  -> stop domain services
  -> close platform/store
  -> flush logger and metrics
  -> return aggregated errors
```

يجب أن يكون `Shutdown` idempotent وآمناً عند الاستدعاء المتكرر والمتزامن.

## 8. الاختبارات المطلوبة

ينبغي أن تغطي اختبارات lifecycle على الأقل:

- configuration غير صالح.
- dependency غير جاهزة.
- فشل bind بسبب استخدام المنفذ.
- عدم إعلان readiness قبل bind.
- إلغاء `Run(ctx)`.
- إيقاف worker طويل والاستدلال على cancellation.
- انتظار workers قبل إغلاق store.
- إلغاء scheduled tasks.
- local Unix socket وcleanup الخاص به.
- concurrent shutdown.
- تجميع أخطاء shutdown.
- فشل startup بعد بدء dependency والتأكد من rollback.
- تشغيل حقيقي يشبه مسار `cmd/server/main.go`.

## 9. التوصيات ذات الأولوية

1. توثيق عقد واضح يحدد هل `Start` يهيئ فقط أم يشغل listener أيضاً.
2. جعل listener وHTTP server تحت ملكية `YouTrackServer` في مسار واحد.
3. ربط local supervision API بالـ entry point عند تفعيلها.
4. إضافة startup guard ورفض البدء المتزامن.
5. إضافة rollback لكل مورد بدأ بنجاح قبل فشل المرحلة التالية.
6. إنشاء registry رسمية للـ workers والـ scheduled jobs.
7. توحيد دلالة `/live` و`/ready` في public وlocal routers.
8. الحفاظ على ترتيب shutdown العكسي للاعتماديات.
9. إضافة اختبار lifecycle كامل يشبه production.
10. ترك WebSocket وcache وsearch وmetrics قدرات اختيارية إلى أن تصبح متطلبات فعلية.

## 10. الحكم النهائي

YouTrack يملك أساساً صحيحاً وقابلاً للتطوير: state machine، readiness صريحة، context cancellation، worker tracking، وshutdown idempotent. الفجوة الرئيسية مع Mattermost ليست في أسماء الدوال، بل في مقدار orchestration الفعلي الذي يملكه الخادم.

أهم ما ينقص YouTrack هو ربط كل مورد يبدأه الخادم بمسار lifecycle كامل: إنشاء، تحقق، تشغيل، مراقبة، إلغاء، انتظار، وإغلاق. عند اكتمال هذا الربط ستصبح طبقة `app` قابلة للتشغيل والاختبار كخادم مستقل، بدلاً من الاعتماد على معرفة خاصة موزعة بين `main` وAPI وChannels.
