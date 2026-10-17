package perms2

// ScopeLevel defines the level at which a permission applies in YouTrack.
type ScopeLevel string

const (
	// ScopeGlobal applies system-wide across all organizations and projects.
	ScopeGlobal ScopeLevel = "GLOBAL"
	// ScopeOrganization applies within a specific organization hierarchy.
	ScopeOrganization ScopeLevel = "ORGANIZATION"
	// ScopeProject applies within a specific project.
	ScopeProject ScopeLevel = "PROJECT"
)

// EntityType represents the target resource being governed.
type EntityType string

const (
	EntitySystem         EntityType = "SYSTEM"
	EntityApp            EntityType = "APP"
	EntityArticle        EntityType = "ARTICLE"
	EntityArticleComment EntityType = "ARTICLE_COMMENT"
	EntityIssue          EntityType = "ISSUE"
	EntityAttachment     EntityType = "ATTACHMENT"
	EntityComment        EntityType = "COMMENT"
	EntityWorkItem       EntityType = "WORK_ITEM"
	EntityOrganization   EntityType = "ORGANIZATION"
	EntityProject        EntityType = "PROJECT"
	EntityUser           EntityType = "USER"
	EntityWatchFolder    EntityType = "WATCH_FOLDER"
)

// OperationType defines the CRUD or operational action.
type OperationType string

const (
	OpCreate  OperationType = "CREATE"
	OpRead    OperationType = "READ"
	OpUpdate  OperationType = "UPDATE"
	OpDelete  OperationType = "DELETE"
	OpLink    OperationType = "LINK"
	OpShare   OperationType = "SHARE"
	OpAdmin   OperationType = "ADMIN"
	OpSpecial OperationType = "SPECIAL"
)

// Permission models a single authorization rule in YouTrack.
type Permission struct {
	ID             PermKey       `json:"id"`
	DisplayName    string        `json:"display_name"`
	Description    string        `json:"description"`
	Entity         EntityType    `json:"entity"`
	Scope          ScopeLevel    `json:"scope"`
	Operation      OperationType `json:"operation"`
	ImpliedPerms   []PermKey     `json:"implied_perms,omitempty"`
	DependentPerms []PermKey     `json:"dependent_perms,omitempty"`
}

// ScopeContext represents the contextual evaluation target.
type ScopeContext struct {
	Type         ScopeLevel `json:"type"`
	ScopeID      string     `json:"scope_id"`
	ResourceType EntityType `json:"resource_type,omitempty"`
}

// RoleAssignment models a user's assigned role with resolved permissions.
// type RoleAssignment struct {
// 	AssignmentID   string       `json:"assignment_id"`
// 	RoleID         string       `json:"role_id"`
// 	RoleName       string       `json:"role_name"`
// 	ScopeType      ScopeLevel   `json:"scope_type"`
// 	OrganizationID string       `json:"organization_id,omitempty"`
// 	ProjectID      string       `json:"project_id,omitempty"`
// 	Permissions    []Permission `json:"permissions"`
// }

// ScopeRole pairs a scope with role information and permissions.
type ScopeRole struct {
	ScopeType   ScopeLevel   `json:"scope_type"`
	RoleName    string       `json:"role_name"`
	Permissions []Permission `json:"permissions"`
}

// InherentAction represents actions that can be performed via inherent rights.
type InherentAction string

const (
	InherentReadOwnIssuePublicFields   InherentAction = "READ_OWN_ISSUE_PUBLIC_FIELDS"
	InherentUpdateOwnIssuePublicFields InherentAction = "UPDATE_OWN_ISSUE_PUBLIC_FIELDS"
	InherentLinkOwnIssue               InherentAction = "LINK_OWN_ISSUE"
	InherentModifyOwnAttachment        InherentAction = "MODIFY_OWN_ATTACHMENT"
	InherentDeleteOwnAttachment        InherentAction = "DELETE_OWN_ATTACHMENT"
	InherentRestrictOwnAttachment      InherentAction = "RESTRICT_OWN_ATTACHMENT"
	InherentReadOwnIssueComment        InherentAction = "READ_OWN_ISSUE_COMMENT"
	InherentReadOwnWorkItem            InherentAction = "READ_OWN_WORK_ITEM"
	InherentReadOwnArticleComment      InherentAction = "READ_OWN_ARTICLE_COMMENT"
	InherentUpdateOwnArticleComment    InherentAction = "UPDATE_OWN_ARTICLE_COMMENT"
	InherentDeleteOwnArticleComment    InherentAction = "DELETE_OWN_ARTICLE_COMMENT"
)

// PermKey identifies a permission in the catalog. It is an alias of string so the
// permission service stays usable from string-based callers (roles_service, JSON payloads).
type PermKey = string

const (
	// System (Global)
	PermLowLevelAdminRead  PermKey = "ADMIN_READ_APP"
	PermLowLevelAdminWrite PermKey = "ADMIN_UPDATE_APP"

	// App (Project)
	PermReadAppContent   PermKey = "READ_APP_CONTENT"
	PermUpdateAppContent PermKey = "UPDATE_APP_CONTENT"

	// Article (Project)
	PermCreateArticle PermKey = "CREATE_ARTICLE"
	PermDeleteArticle PermKey = "DELETE_ARTICLE"
	PermReadArticle   PermKey = "READ_ARTICLE"
	PermUpdateArticle PermKey = "UPDATE_ARTICLE"

	// Article Comment (Project)
	PermCreateArticleComment PermKey = "CREATE_ARTICLE_COMMENT"
	PermDeleteArticleComment PermKey = "DELETE_ARTICLE_COMMENT"
	PermReadArticleComment   PermKey = "READ_ARTICLE_COMMENT"
	PermUpdateArticleComment PermKey = "UPDATE_ARTICLE_COMMENT"

	// Issue (Project)
	PermApplyCommandsSilently    PermKey = "APPLY_COMMANDS_SILENTLY"
	PermCreateIssue              PermKey = "CREATE_ISSUE"
	PermDeleteIssue              PermKey = "DELETE_ISSUE"
	PermLinkIssue                PermKey = "LINK_ISSUE"
	PermOverrideVisibility       PermKey = "READ_HIDDEN_STUFF"
	PermReadIssue                PermKey = "READ_ISSUE"
	PermReadIssuePrivateFields   PermKey = "PRIVATE_READ_ISSUE"
	PermUpdateIssue              PermKey = "UPDATE_ISSUE"
	PermUpdateIssuePrivateFields PermKey = "PRIVATE_UPDATE_ISSUE"
	PermUpdateWatchers           PermKey = "UPDATE_WATCHERS"
	PermViewVoters               PermKey = "VIEW_VOTERS"
	PermViewWatchers             PermKey = "VIEW_WATCHERS"

	// Issue Attachment (Project)
	PermAddAttachment    PermKey = "CREATE_ATTACHMENT_ISSUE"
	PermDeleteAttachment PermKey = "DELETE_ATTACHMENT_ISSUE"
	PermUpdateAttachment PermKey = "UPDATE_ATTACHMENT_ISSUE"

	// Issue Comment (Project)
	PermCreateIssueComment                    PermKey = "CREATE_COMMENT"
	PermDeleteIssueComment                    PermKey = "DELETE_COMMENT"
	PermDeleteNotOwnAndPermanentCommentDelete PermKey = "DELETE_NOT_OWN_COMMENT"
	PermReadIssueComment                      PermKey = "READ_COMMENT"
	PermUpdateIssueComment                    PermKey = "UPDATE_COMMENT"
	PermUpdateNotOwnIssueComment              PermKey = "UPDATE_NOT_OWN_COMMENT"

	// Issue Work Item (Project)
	PermCreateNotOwnWorkItem PermKey = "CREATE_NOT_OWN_WORK_ITEM"
	PermCreateWorkItem       PermKey = "CREATE_WORK_ITEM"
	PermReadWorkItem         PermKey = "READ_WORK_ITEM"
	PermUpdateNotOwnWorkItem PermKey = "UPDATE_NOT_OWN_WORK_ITEM"
	PermUpdateWorkItem       PermKey = "UPDATE_WORK_ITEM"

	// Organization (Global / Organization)
	PermCreateOrganization PermKey = "CREATE_ORGANIZATION"
	PermDeleteOrganization PermKey = "DELETE_ORGANIZATION"
	PermReadOrganization   PermKey = "READ_ORGANIZATION"
	PermUpdateOrganization PermKey = "UPDATE_ORGANIZATION"

	// Project (Global / Project)
	PermCreateProject    PermKey = "CREATE_PROJECT"
	PermDeleteProject    PermKey = "DELETE_PROJECT"
	PermReadProjectBasic PermKey = "READ_PROJECT_BASIC"
	PermReadProjectFull  PermKey = "READ_PROJECT"
	PermUpdateProject    PermKey = "UPDATE_PROJECT"

	// Users (Global)
	PermCreateUser      PermKey = "CREATE_USER"
	PermDeleteUser      PermKey = "DELETE_USER"
	PermReadUserBasic   PermKey = "READ_USER_BASIC"
	PermReadUserDetails PermKey = "READ_USER"
	PermUpdateSelf      PermKey = "UPDATE_PROFILE"
	PermUpdateUser      PermKey = "UPDATE_USER"

	// Watch Folder / Saved Searches & Tags
	PermCreateWatchFolder PermKey = "CREATE_WATCH_FOLDER"
	PermDeleteWatchFolder PermKey = "DELETE_WATCH_FOLDER"
	PermUpdateWatchFolder PermKey = "UPDATE_WATCH_FOLDER"
	PermShareWatchFolder  PermKey = "SHARE_WATCH_FOLDER"
)

var permissionsNames []PermKey = []PermKey{
	// System (Global)
	PermLowLevelAdminRead,
	PermLowLevelAdminWrite,

	// App (Project)
	PermReadAppContent,
	PermUpdateAppContent,

	// Article (Project)
	PermCreateArticle,
	PermDeleteArticle,
	PermReadArticle,
	PermUpdateArticle,

	// Article Comment (Project)
	PermCreateArticleComment,
	PermDeleteArticleComment,
	PermReadArticleComment,
	PermUpdateArticleComment,

	// Issue (Project)
	PermApplyCommandsSilently,
	PermCreateIssue,
	PermDeleteIssue,
	PermLinkIssue,
	PermOverrideVisibility,
	PermReadIssue,
	PermReadIssuePrivateFields,
	PermUpdateIssue,
	PermUpdateIssuePrivateFields,
	PermUpdateWatchers,
	PermViewVoters,
	PermViewWatchers,

	// Issue Attachment (Project)
	PermAddAttachment,
	PermDeleteAttachment,
	PermUpdateAttachment,

	// Issue Comment (Project)
	PermCreateIssueComment,
	PermDeleteIssueComment,
	PermDeleteNotOwnAndPermanentCommentDelete,
	PermReadIssueComment,
	PermUpdateIssueComment,
	PermUpdateNotOwnIssueComment,

	// Issue Work Item (Project)
	PermCreateNotOwnWorkItem,
	PermCreateWorkItem,
	PermReadWorkItem,
	PermUpdateNotOwnWorkItem,
	PermUpdateWorkItem,

	// Organization (Global / Organization)
	PermCreateOrganization,
	PermDeleteOrganization,
	PermReadOrganization,
	PermUpdateOrganization,

	// Project (Global / Project)
	PermCreateProject,
	PermDeleteProject,
	PermReadProjectBasic,
	PermReadProjectFull,
	PermUpdateProject,

	// Users (Global)
	PermCreateUser,
	PermDeleteUser,
	PermReadUserBasic,
	PermReadUserDetails,
	PermUpdateSelf,
	PermUpdateUser,

	// Watch Folder / Saved Searches & Tags
	PermCreateWatchFolder,
	PermDeleteWatchFolder,
	PermUpdateWatchFolder,
	PermShareWatchFolder,
}
