package domain

// Permission يمثل الإذن أو الصلاحية في YouTrack
type Permission struct {
	ID                             string `json:"id"`
	Name                           string `json:"name,omitempty"`
	Description                    string `json:"description,omitempty"`
	PermissionEntityType           string `json:"permissionEntityType,omitempty"`
	LocalizedPermissionEntityType  string `json:"localizedPermissionEntityType,omitempty"`
	Operation                      string `json:"operation,omitempty"`
	IsGlobal                       bool   `json:"isGlobal,omitempty"`
	Type                           string `json:"$type,omitempty"`
}

// Role يمثل الدور الذي يجمع عدة صلاحيات
type Role struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	Description   string        `json:"description,omitempty"`
	AuditTargetID string        `json:"auditTargetId,omitempty"`
	IsUpdatable   bool          `json:"isUpdatable,omitempty"`
	Immutable     bool          `json:"immutable,omitempty"`
	Permissions   []*Permission `json:"permissions,omitempty"`
	Type          string        `json:"$type,omitempty"`
}

// AssignedRole يمثل تعيين دور معين لمستخدم أو فريق في نطاق معين
type AssignedRole struct {
	ID            string      `json:"id"`
	Role          *Role       `json:"role,omitempty"`
	AuditTargetID string      `json:"auditTargetId,omitempty"`
	Holder        any         `json:"holder,omitempty"` // User or ProjectTeam or UserGroup
	Scope         *Scope      `json:"scope,omitempty"`
	Type          string      `json:"$type,omitempty"`
}

// Scope يمثل النطاق الجغرافي للدور (مثل مشروع محدد أو عام)
type Scope struct {
	ID           string        `json:"id"`
	Organization *Organization `json:"organization,omitempty"`
	Project      *Project      `json:"project,omitempty"`
	Type         string        `json:"$type,omitempty"`
}
