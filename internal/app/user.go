package app

import (
	"context"

	"youtrack_backend/internal/api/fields"
	"youtrack_backend/internal/model"
)

// GetUserByID يعيد مستخدمًا بمعرّفه.
func (a *App) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	if id == "me" {
		return nil, model.Unauthorized("user context missing")
	}
	u, err := a.store.Users().GetByID(ctx, id)
	if err != nil {
		return nil, model.NotFound("user %q not found", id)
	}
	u.Normalize()
	return u, nil
}

// GetCurrentUser يعيد المستخدم الحالي بمعرّفه مع كامل مرفقاته بناءً على شجرة الحقول.
func (a *App) GetCurrentUser(ctx context.Context, userID string, tree *fields.FieldTree) (*model.User, error) {
	u, err := a.store.Users().GetMe(ctx, userID, tree)
	if err != nil {
		return nil, model.NotFound("user not found")
	}
	return u, nil
}

// GetAllUsers يعيد قائمة المستخدمين.
func (a *App) GetAllUsers(ctx context.Context) ([]*model.User, error) {
	users, err := a.store.Users().All(ctx)
	if err != nil {
		return nil, model.Internal("failed to load users: %v", err)
	}
	for _, u := range users {
		u.Normalize()
	}
	return users, nil
}

// GetGrazieProfile يعيد إعدادات Grazie للمستخدم.
func (a *App) GetGrazieProfile(ctx context.Context, userID string) (*model.GrazieUserProfile, error) {
	return a.store.Users().GetGrazieProfile(ctx, userID)
}

// GetGeneralProfile يعيد الإعدادات العامة لملف المستخدم.
func (a *App) GetGeneralProfile(ctx context.Context, userID string) (*model.GeneralUserProfile, error) {
	return a.store.Users().GetGeneralProfile(ctx, userID)
}

// GetQuestionnaireProfile يعيد إعدادات الاستبيانات للمستخدم.
func (a *App) GetQuestionnaireProfile(ctx context.Context, userID string) (*model.QuestionnaireUserProfile, error) {
	return a.store.Users().GetQuestionnaireProfile(ctx, userID)
}

// GetRecentIssues يعيد المشاكل المشاهدة مؤخرًا للمستخدم.
func (a *App) GetRecentIssues(ctx context.Context, userID string, top int, skip int) ([]*model.RecentIssue, error) {
	return a.store.Users().GetRecentIssues(ctx, userID, top, skip)
}

// GetRecentArticles يعيد المقالات المشاهدة مؤخرًا للمستخدم.
func (a *App) GetRecentArticles(ctx context.Context, userID string, top int, skip int) ([]*model.RecentArticle, error) {
	return a.store.Users().GetRecentArticles(ctx, userID, top, skip)
}

// GetHubCurrentUser يعيد ملف المستخدم في Hub.
func (a *App) GetHubCurrentUser(ctx context.Context, userID string) (*model.HubUser, error) {
	return a.store.Users().GetHubMe(ctx, userID)
}

// GetInboxFolders يعيد مجلدات صندوق الوارد للمستخدم (مطابق لـ request8.txt).
func (a *App) GetInboxFolders(ctx context.Context, userID string) ([]*model.InboxFolder, error) {
	return a.store.Users().InboxFolders(ctx, userID)
}

// GetInboxThreads يعيد خيوط الرسائل للمستخدم مع دعم الجلب الانتقائي (مطابق لـ request9.txt).
func (a *App) GetInboxThreads(ctx context.Context, userID string, top int, skip int, tree *fields.FieldTree) ([]*model.InboxThread, error) {
	return a.store.Inbox().Threads(ctx, userID, top, skip, tree)
}
