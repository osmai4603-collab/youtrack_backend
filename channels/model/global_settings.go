package model

// RestCorsSettings يمثّل إعدادات CORS الخاصة بواجهة REST
// (مطابق لـ $type: "RestCorsSettings" في استجابة YouTrack).
type RestCorsSettings struct {
	AllowAllOrigins bool     `json:"allowAllOrigins"`
	AllowedOrigins  []string `json:"allowedOrigins"`
	Type            string   `json:"$type,omitempty"`
}

// ImageTextRecognitionSettings يمثّل إعدادات التعرف على النص في الصور (OCR)
// (مطابق لـ $type: "ImageTextRecognitionSettings").
type ImageTextRecognitionSettings struct {
	Enabled bool   `json:"enabled"`
	Type    string `json:"$type,omitempty"`
}

// SystemSettings يمثّل إعدادات النظام العامة (OCR مدعوم أم لا)
// (مطابق لـ $type: "SystemSettings").
type SystemSettings struct {
	OcrSupported string `json:"ocrSupported"`
	Type         string `json:"$type,omitempty"`
}

// EmailSettings يمثّل إعدادات البريد الإلكتروني للإشعارات
// (مطابق لـ $type: "EmailSettings").
type EmailSettings struct {
	IsEnabled bool   `json:"isEnabled"`
	IsDefault bool   `json:"isDefault"`
	Type      string `json:"$type,omitempty"`
}

// NotificationSettings يمثّل إعدادات الإشعارات العامة
// (مطابق لـ $type: "NotificationSettings").
type NotificationSettings struct {
	EmailSettings *EmailSettings `json:"emailSettings"`
	Type          string         `json:"$type,omitempty"`
}

// AdminGlobalSettings يمثّل المخطط الجديد المستقل الخاص بنقطة
// GET /api/admin/globalSettings (مطابق لـ request5.txt و request19.txt و request46.txt).
// الاسم مختلف عن GlobalSettings القائم في config.go ليُبقى ذلك المخطط دون أي تعديل.
// قيمته في الـ JSON هي "$type": "GlobalSettings" كما في استجابة YouTrack الأصلية.
type AdminGlobalSettings struct {
	RestSettings                 *RestCorsSettings             `json:"restSettings,omitempty"`
	ImageTextRecognitionSettings *ImageTextRecognitionSettings `json:"imageTextRecognitionSettings,omitempty"`
	SystemSettings               *SystemSettings               `json:"systemSettings,omitempty"`
	NotificationSettings         *NotificationSettings         `json:"notificationSettings,omitempty"`
	Type                         string                        `json:"$type,omitempty"`
}

// DefaultAdminGlobalSettings يعيد القيم الافتراضية القياسية الكاملة المتوافقة مع YouTrack.
func DefaultAdminGlobalSettings() *AdminGlobalSettings {
	return &AdminGlobalSettings{
		Type: "GlobalSettings",
		RestSettings: &RestCorsSettings{
			AllowAllOrigins: false,
			AllowedOrigins:  []string{},
			Type:            "RestCorsSettings",
		},
		ImageTextRecognitionSettings: &ImageTextRecognitionSettings{
			Enabled: true,
			Type:    "ImageTextRecognitionSettings",
		},
		SystemSettings: &SystemSettings{
			OcrSupported: "true",
			Type:         "SystemSettings",
		},
		NotificationSettings: &NotificationSettings{
			EmailSettings: &EmailSettings{
				IsEnabled: true,
				IsDefault: true,
				Type:      "EmailSettings",
			},
			Type: "NotificationSettings",
		},
	}
}
