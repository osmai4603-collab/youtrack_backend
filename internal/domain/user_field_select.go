package domain

// UserFieldSelect يحمل الحقائق/الأقسام المطلوبة من نقطة /api/users/me كما استُخرجت
// من معامل fields في Query String. يُعبَّر عن كل قسم بعلامة منطقية (true = مطلوب)،
// بحيث يُبنى طلب قاعدة البيانات على الأقسام المطلوبة فقط بدلاً من جلب مخطط المستخدم
// كاملاً (بجميع إعداداته) ثم تصفيته لاحقاً.
//
// إذا كان الكائن nil فهذا يعني "جلب كل شيء" (لم يُمرَّر معامل fields).
type UserFieldSelect struct {
	// أقسام المستوى الجذري (غير الرخيصة / المحتمل أنها ثقيلة)
	Profiles         bool
	Widgets          bool
	FeatureFlags     bool
	IssueRelatedGroup bool

	// الأقسام الفرعية لملف profiles
	General       bool
	Articles      bool
	TimeTracking  bool
	Tips          bool
	Appearance    bool
	IssuesList    bool
	Helpdesk      bool
	AI            bool
	Notifications bool
}

// AnyProfiles يتحقق ما إذا كان أي قسم من ملف التفضيلات (profiles) مطلوباً.
func (s *UserFieldSelect) AnyProfiles() bool {
	return s.Profiles || s.General || s.Articles || s.TimeTracking || s.Tips ||
		s.Appearance || s.IssuesList || s.Helpdesk || s.AI || s.Notifications
}

// AnyRootSection يتحقق ما إذا كان أي قسم ثقيل في المستوى الجذري مطلوباً
// (profiles, widgets, featureFlags, issueRelatedGroup).
func (s *UserFieldSelect) AnyRootSection() bool {
	return s.Profiles || s.Widgets || s.FeatureFlags || s.IssueRelatedGroup
}
