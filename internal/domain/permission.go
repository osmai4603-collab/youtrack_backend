package domain

// CachedPermission يمثل الصلاحية المخزنة مؤقتاً للمستخدم مع نطاق تطبيقها
type CachedPermission struct {
	ID            string               `json:"id"`
	Global        bool                 `json:"global"`
	Projects      []*PermissionProject `json:"projects,omitempty"`
	Organizations []*Organization      `json:"organizations,omitempty"`
	Permission    *Permission          `json:"permission,omitempty"`
	Type          string               `json:"$type,omitempty"`
}

// PermissionProject يمثل المشروع المرتبط بصلاحية مخزنة
type PermissionProject struct {
	ID          string       `json:"id"`
	ProjectType *ProjectType `json:"projectType,omitempty"`
	Type        string       `json:"$type,omitempty"`
}
