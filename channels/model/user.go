package model

// UserType يصف نوع المستخدم (Standard, Agent, Reporter, ...)
type UserType struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"$type"`
}

// UserGroup يمثّل مجموعة مستخدمين في النظام.
type UserGroup struct {
	ID            string `json:"id" db:"id"`
	Name          string `json:"name" db:"name"`
	GroupType     string `json:"groupType,omitempty" db:"group_type"`
	AllUsersGroup bool   `json:"allUsersGroup" db:"all_users_group"`
	Icon          string `json:"icon,omitempty" db:"icon"`
	Description   string `json:"description,omitempty" db:"description"`
	AuditTargetID string `json:"auditTargetId,omitempty" db:"audit_target_id"`
	IsUpdatable   bool   `json:"isUpdatable" db:"is_updatable"`
	IsRemovable   bool   `json:"isRemovable" db:"is_removable"`
	TeamForProject *ProjectRef `json:"teamForProject,omitempty"`
	Type          string `json:"$type,omitempty"`
}

// IssueRelatedGroup يصف المجموعات المرتبطة بالقضايا.
type IssueRelatedGroup struct {
	PermittedGroups []*UserGroup `json:"permittedGroups,omitempty"`
	Type            string       `json:"$type,omitempty"`
}

// User يصف مستخدم النظام مع كامل مرفقاته وإعداداته.
// يتوافق الحقل مع جدول users في قاعدة البيانات عبر وسم db.
type User struct {
	ID                string             `json:"id" db:"id"`
	Login             string             `json:"login" db:"login"`
	Email             string             `json:"email,omitempty" db:"email"`
	FullName          string             `json:"fullName,omitempty" db:"full_name"`
	Name              string             `json:"name,omitempty" db:"name"`
	AvatarURL         string             `json:"avatarUrl,omitempty" db:"avatar_url"`
	UserTypeID        string             `json:"-" db:"user_type_id"`
	UserType          *UserType          `json:"userType,omitempty"`
	IsEmailVerified   bool               `json:"isEmailVerified" db:"is_email_verified"`
	Guest             bool               `json:"guest" db:"guest"`
	Online            bool               `json:"online" db:"online"`
	Banned            bool               `json:"banned" db:"banned"`
	BanBadge          *string            `json:"banBadge" db:"ban_badge"`
	CanReadProfile    bool               `json:"canReadProfile" db:"can_read_profile"`
	IsLocked          bool               `json:"isLocked" db:"is_locked"`
	RingID            string             `json:"ringId,omitempty" db:"ring_id"`
	PasswordHash      string             `json:"-" db:"password_hash"`
	IssueRelatedGroup *IssueRelatedGroup `json:"issueRelatedGroup,omitempty"`
	FeatureFlags      []*FeatureFlag     `json:"featureFlags,omitempty"`
	Widgets           []*DashboardWidget `json:"widgets,omitempty"`
	Profiles          *UserProfiles      `json:"profiles,omitempty"`
	Type              string             `json:"$type,omitempty"`
}

// Normalize يملأ الحقول المحسوبة لمستخدم عام.
func (u *User) Normalize() {
	if u.Type == "" {
		u.Type = "User"
	}
	if u.UserType != nil && u.UserType.Type == "" {
		u.UserType.Type = "UserType"
	}
	if u.Widgets == nil {
		u.Widgets = []*DashboardWidget{}
	}
}

// NormalizeMe يملأ الحقول للمستخدم الحالي ($type: "Me").
func (u *User) NormalizeMe() {
	u.Type = "Me"
	if u.UserType != nil && u.UserType.Type == "" {
		u.UserType.Type = "UserType"
	}
	if u.Widgets == nil {
		u.Widgets = []*DashboardWidget{}
	}
}

// UserProfile يصف إعدادات واجهة المستخدم الأساسية للتوافق.
type UserProfile struct {
	UserID                  string `json:"userId" db:"user_id"`
	TimezoneID              string `json:"timezoneId,omitempty" db:"timezone_id"`
	LocaleID                string `json:"localeId,omitempty" db:"locale_id"`
	DatePattern             string `json:"datePattern,omitempty" db:"date_pattern"`
	DateFieldPattern        string `json:"dateFieldPattern,omitempty" db:"date_field_pattern"`
	EmailNotifications      bool   `json:"emailNotificationsEnabled,omitempty" db:"email_notifications_enabled"`
	MentionNotifications    bool   `json:"mentionNotificationsEnabled,omitempty" db:"mention_notifications_enabled"`
	CompactMode             bool   `json:"compactMode,omitempty" db:"compact_mode"`
	ExpandNavigation        bool   `json:"expandNavigation,omitempty" db:"expand_navigation"`
	IsTimeTrackingAvailable bool   `json:"isTimeTrackingAvailable,omitempty" db:"is_time_tracking_available"`
}

