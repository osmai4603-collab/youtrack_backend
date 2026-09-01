package model

// SecurityFilterField يمثّل حقل تصفية في بحث الأمان.
type SecurityFilterField struct {
	ID         string `json:"id"`
	Name       string `json:"name,omitempty"`
	EntityType string `json:"-"` // يستخدم للتصفية داخلياً فقط
	Type       string `json:"$type,omitempty"`
}
