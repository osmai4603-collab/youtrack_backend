package domain

// SecurityFilterField يمثل حقل البحث الأمني للمستخدمين والأدوار
type SecurityFilterField struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"$type,omitempty"`
}
