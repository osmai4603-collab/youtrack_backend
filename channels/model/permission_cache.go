package model

// PermissionCacheEntry يمثّل صلاحية مخبأة لاستجابة GET /api/permissions/cache
// مع دعم الجلب الانتقائي عبر معامل fields (مطابق لـ request20.txt).
type PermissionCacheEntry struct {
	ID             string                          `json:"id,omitempty"`
	Global         *bool                           `json:"global,omitempty"`
	Projects       []*PermissionCacheProject       `json:"projects"`
	Organizations  []*PermissionCacheOrganization  `json:"organizations"`
	Type           string                          `json:"$type,omitempty"`
	PermissionName string                          `json:"-"` // حقل داخلي لدعم Request 26
}

// PermissionCacheProject يمثّل مشروعاً داخل صلاحية مخبأة.
type PermissionCacheProject struct {
	ID          string       `json:"id,omitempty"`
	ProjectType *ProjectType `json:"projectType,omitempty"`
	Type        string       `json:"$type,omitempty"`
}

// PermissionCacheOrganization يمثّل منظمة داخل صلاحية مخبأة.
type PermissionCacheOrganization struct {
	ID   string `json:"id,omitempty"`
	Type string `json:"$type,omitempty"`
}
