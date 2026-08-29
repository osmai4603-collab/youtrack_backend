package store

import (
	"context"

	"youtrack_backend/internal/api/fields"
	"youtrack_backend/internal/model"
)

// Store يجمّع كل المستودعات الفرعية.
type Store interface {
	Users() UserStore
	Projects() ProjectStore
	Issues() IssueStore
	Admin() AdminStore
	Inbox() InboxStore
	SavedQueries() SavedQueryStore
	Search() SearchStore
}

// SearchStore يحدّد عمليات البحث ومساعد البحث.
type SearchStore interface {
	GetAssist(ctx context.Context, query string, caret int, tree *fields.FieldTree) (*model.SearchAssistResponse, error)
}

// SavedQueryStore يحدّد عمليات الاستعلامات المحفوظة.
type SavedQueryStore interface {
	SavedQueries(ctx context.Context, tree *fields.FieldTree, top int, skip int) ([]*model.SavedQuery, error)
}

// UserStore يحدّد عمليات المستخدمين المتاحة.
type UserStore interface {
	GetByID(ctx context.Context, id string) (*model.User, error)
	GetByLogin(ctx context.Context, login string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	Create(ctx context.Context, u *model.User) error
	All(ctx context.Context) ([]*model.User, error)
	GetProfile(ctx context.Context, userID string) (*model.UserProfile, error)
	CreateProfile(ctx context.Context, p *model.UserProfile) error
	GetMe(ctx context.Context, userID string, tree *fields.FieldTree) (*model.User, error)
	GetFeatureFlags(ctx context.Context) ([]*model.FeatureFlag, error)
	GetRecentIssues(ctx context.Context, userID string, limit int, offset int) ([]*model.RecentIssue, error)
	GetRecentArticles(ctx context.Context, userID string, limit int, offset int) ([]*model.RecentArticle, error)
	GetGrazieProfile(ctx context.Context, userID string) (*model.GrazieUserProfile, error)
	GetGeneralProfile(ctx context.Context, userID string) (*model.GeneralUserProfile, error)
	GetQuestionnaireProfile(ctx context.Context, userID string) (*model.QuestionnaireUserProfile, error)
	GetHubMe(ctx context.Context, userID string) (*model.HubUser, error)
	InboxFolders(ctx context.Context, userID string) ([]*model.InboxFolder, error)
}

// ProjectStore يحدّد عمليات المشاريع المتاحة.
type ProjectStore interface {
	GetByID(ctx context.Context, id string) (*model.Project, error)
	GetByShortName(ctx context.Context, shortName string) (*model.Project, error)
	GetDetailed(ctx context.Context, id string, tree *fields.FieldTree) (*model.Project, error)
	All(ctx context.Context) ([]*model.Project, error)
}

// IssueStore يحدّد عمليات القضايا المتاحة.
type IssueStore interface {
	GetByID(ctx context.Context, id string) (*model.Issue, error)
	GetByReadableID(ctx context.Context, idReadable string) (*model.Issue, error)
	All(ctx context.Context, query string, limit int) ([]*model.Issue, error)
	Create(ctx context.Context, i *model.Issue) error
	Update(ctx context.Context, i *model.Issue) error
	Delete(ctx context.Context, id string) error
	Comments(ctx context.Context, issueID string) ([]*model.IssueComment, error)
	CreateComment(ctx context.Context, c *model.IssueComment) error
	Tags(ctx context.Context, issueID string) ([]*model.Tag, error)
	Links(ctx context.Context, issueID string) ([]*model.IssueLink, error)
	GetSortedIssues(ctx context.Context, folderID string, query string, top int, skip int) ([]*model.IssueTreeItem, error)
}

// AdminStore يحدّد عمليات الإدارة والميتاداتا.
type AdminStore interface {
	Roles(ctx context.Context) ([]*model.Role, error)
	Permissions(ctx context.Context) ([]*model.Permission, error)
	GlobalSettings(ctx context.Context) (*model.GlobalSettings, error)
	AdminGlobalSettings(ctx context.Context, tree *fields.FieldTree) (*model.AdminGlobalSettings, error)
	BannersConfig(ctx context.Context, tree *fields.FieldTree) (*model.BannersConfig, error)
	Widgets(ctx context.Context) ([]*model.DashboardWidget, error)
	CachedPermissions(ctx context.Context, userID string) ([]*model.CachedPermission, error)
	ProjectDashboard(ctx context.Context, projectKey string, tree *fields.FieldTree) (*model.ProjectDashboard, error)
}

// InboxStore يحدّد عمليات صندوق الوارد.
type InboxStore interface {
	Threads(ctx context.Context, userID string, top int, skip int, tree *fields.FieldTree) ([]*model.InboxThread, error)
}
