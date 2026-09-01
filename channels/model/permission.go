package model

// Permission يمثّل صلاحية واحدة في النظام.
type Permission struct {
	ID                            string `json:"id" db:"id"`
	Name                          string `json:"name" db:"name"`
	Description                   string `json:"description,omitempty" db:"description"`
	PermissionEntityType          string `json:"permissionEntityType,omitempty" db:"permission_entity_type"`
	LocalizedPermissionEntityType string `json:"localizedPermissionEntityType,omitempty" db:"localized_permission_entity_type"`
	Operation                     string `json:"operation,omitempty" db:"operation"`
	IsGlobal                      bool   `json:"global,omitempty" db:"is_global"`
	Type                          string `json:"$type,omitempty"`
}

// Role يمثّل دورًا يحوي مجموعة صلاحيات.
type Role struct {
	ID          string        `json:"id" db:"id"`
	Name        string        `json:"name" db:"name"`
	Description string        `json:"description,omitempty" db:"description"`
	IsUpdatable bool          `json:"isUpdatable,omitempty" db:"is_updatable"`
	Immutable   bool          `json:"immutable,omitempty" db:"immutable"`
	Permissions []*Permission `json:"permissions,omitempty"`
	Type        string        `json:"$type,omitempty"`
}

// RolePermission يربط دورًا بصلاحية.
type RolePermission struct {
	RoleID       string `json:"roleId" db:"role_id"`
	PermissionID string `json:"permissionId" db:"permission_id"`
}

// CachedPermissionProject يمثّل مشروعًا داخل صلاحية مخبأة.
type CachedPermissionProject struct {
	ProjectType *ProjectType `json:"projectType"`
	ID          string       `json:"id"`
	Type        string       `json:"$type"`
}

// CachedPermission يمثّل صلاحية مخبأة للمستخدم مع نطاقاتها (مطابق لـ request7.txt).
type CachedPermission struct {
	ID            string                     `json:"id" db:"id"`
	Global        bool                       `json:"global" db:"is_global"`
	Projects      []*CachedPermissionProject `json:"projects"`
	Organizations []*Organization            `json:"organizations"`
	Type          string                     `json:"$type"`
}
