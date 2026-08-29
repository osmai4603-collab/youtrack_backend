package model

// SystemEventsBanner يمثّل لافتة حدث نظام واحدة داخل banners
// (قيمتها في الـ JSON هي "$type": "SystemEventsBanner" كما في استجابة YouTrack).
type SystemEventsBanner struct {
	ID   int    `json:"id"`
	Text string `json:"text,omitempty"`
	Type string `json:"$type,omitempty"`
}

// BannersConfig يمثّل المخطط الجديد المستقل الخاص بكائن banners لنقطة
// GET /api/config (مطابق لـ request11.txt).
// اسمه مختلف عن أي مخطط قائم لِيُبقى كل المخططات الأخرى دون أي تعديل.
// قيمته في الـ JSON هي "$type": "BannersConfig" كما في استجابة YouTrack الأصلية.
type BannersConfig struct {
	GlobalBanner        string                `json:"globalBanner,omitempty"`
	GlobalBannerEnabled bool                  `json:"globalBannerEnabled"`
	SystemEventsBanners []*SystemEventsBanner `json:"systemEventsBanners"`
	Type                string                `json:"$type,omitempty"`
}

// DefaultBannersConfig يعيد القيم الافتراضية القياسية الكاملة المتوافقة مع YouTrack
// (المطابقة لاستجابة request11.txt).
func DefaultBannersConfig() *BannersConfig {
	return &BannersConfig{
		GlobalBanner:        "An upgrade is currently in progress. YouTrack is available in read-only mode. We estimate that the service will be fully available at 00:42 UTC",
		GlobalBannerEnabled: false,
		SystemEventsBanners: []*SystemEventsBanner{},
		Type:                "BannersConfig",
	}
}
