package model

// PermissionR26 يمثّل صلاحية مخصصة للطلب رقم 26.
type PermissionR26 struct {
	ID   string `json:"id,omitempty"`
	Key  string `json:"key,omitempty"`
	Name string `json:"name,omitempty"`
	Type string `json:"$type,omitempty"`
}

// PermissionCacheEntryR26 يمثّل مدخل ذاكرة تخزين الصلاحيات المخصص للطلب رقم 26.
type PermissionCacheEntryR26 struct {
	Global     *bool                      `json:"global,omitempty"`
	Permission *PermissionR26             `json:"permission,omitempty"`
	Projects   []*PermissionCacheProjectR26 `json:"projects,omitempty"`
	Type       string                     `json:"$type,omitempty"`
}

// PermissionCacheProjectR26 يمثّل مشروعاً داخل صلاحية مخبأة للطلب رقم 26.
type PermissionCacheProjectR26 struct {
	ID   string `json:"id,omitempty"`
	Type string `json:"$type,omitempty"`
}
