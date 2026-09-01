package app

import (
	"youtrack_backend/channels/model"
	"youtrack_backend/channels/model/fields"
	"youtrack_backend/channels/request"
)

// GetUserByID يعيد مستخدمًا بمعرّفه.
func (a *YouTrackApp) GetUserByID(c request.CTX, id string) (*model.User, *model.AppError) {
	if id == "me" {
		return nil, model.NewUnauthorizedError("App.GetUserByID", "user context missing")
	}
	ctx := c.Context()
	u, err := a.Store().Users().GetByID(ctx, id)
	if err != nil {
		return nil, model.NewNotFoundError("App.GetUserByID", "user not found")
	}
	u.Normalize()
	return u, nil
}

// GetCurrentUser يعيد المستخدم الحالي بمعرّفه مع كامل مرفقاته بناءً على شجرة الحقول.
func (a *YouTrackApp) GetCurrentUser(c request.CTX, userID string, tree *fields.FieldTree) (*model.User, *model.AppError) {
	if userID == "" {
		userID = c.UserID()
	}
	ctx := c.Context()
	u, err := a.Store().Users().GetMe(ctx, userID, tree)
	if err != nil {
		return nil, model.NewNotFoundError("App.GetCurrentUser", "user not found")
	}
	return u, nil
}

// GetAllUsers يعيد قائمة المستخدمين.
func (a *YouTrackApp) GetAllUsers(c request.CTX) ([]*model.User, *model.AppError) {
	ctx := c.Context()
	users, err := a.Store().Users().All(ctx)
	if err != nil {
		return nil, model.NewInternalError("App.GetAllUsers", "failed to load users", err)
	}
	for _, u := range users {
		u.Normalize()
	}
	return users, nil
}

// GetGrazieProfile يعيد إعدادات Grazie للمستخدم.
func (a *YouTrackApp) GetGrazieProfile(c request.CTX, userID string) (*model.GrazieUserProfile, *model.AppError) {
	if userID == "" {
		userID = c.UserID()
	}
	ctx := c.Context()
	profile, err := a.Store().Users().GetGrazieProfile(ctx, userID)
	if err != nil {
		return nil, model.NewInternalError("App.GetGrazieProfile", "failed to load grazie profile", err)
	}
	return profile, nil
}

// GetGeneralProfile يعيد الإعدادات العامة لملف المستخدم.
func (a *YouTrackApp) GetGeneralProfile(c request.CTX, userID string) (*model.GeneralUserProfile, *model.AppError) {
	if userID == "" {
		userID = c.UserID()
	}
	ctx := c.Context()
	profile, err := a.Store().Users().GetGeneralProfile(ctx, userID)
	if err != nil {
		return nil, model.NewInternalError("App.GetGeneralProfile", "failed to load general profile", err)
	}
	return profile, nil
}

// GetQuestionnaireProfile يعيد إعدادات الاستبيانات للمستخدم.
func (a *YouTrackApp) GetQuestionnaireProfile(c request.CTX, userID string) (*model.QuestionnaireUserProfile, *model.AppError) {
	if userID == "" {
		userID = c.UserID()
	}
	ctx := c.Context()
	profile, err := a.Store().Users().GetQuestionnaireProfile(ctx, userID)
	if err != nil {
		return nil, model.NewInternalError("App.GetQuestionnaireProfile", "failed to load questionnaire profile", err)
	}
	return profile, nil
}

// GetRecentIssues يعيد المشاكل المشاهدة مؤخرًا للمستخدم.
func (a *YouTrackApp) GetRecentIssues(c request.CTX, userID string, top int, skip int) ([]*model.RecentIssue, *model.AppError) {
	if userID == "" {
		userID = c.UserID()
	}
	ctx := c.Context()
	issues, err := a.Store().Users().GetRecentIssues(ctx, userID, top, skip)
	if err != nil {
		return nil, model.NewInternalError("App.GetRecentIssues", "failed to load recent issues", err)
	}
	return issues, nil
}

// GetRecentArticles يعيد المقالات المشاهدة مؤخرًا للمستخدم.
func (a *YouTrackApp) GetRecentArticles(c request.CTX, userID string, top int, skip int) ([]*model.RecentArticle, *model.AppError) {
	if userID == "" {
		userID = c.UserID()
	}
	ctx := c.Context()
	articles, err := a.Store().Users().GetRecentArticles(ctx, userID, top, skip)
	if err != nil {
		return nil, model.NewInternalError("App.GetRecentArticles", "failed to load recent articles", err)
	}
	return articles, nil
}

// GetHubCurrentUser يعيد ملف المستخدم في Hub.
func (a *YouTrackApp) GetHubCurrentUser(c request.CTX, userID string) (*model.HubUser, *model.AppError) {
	if userID == "" {
		userID = c.UserID()
	}
	ctx := c.Context()
	user, err := a.Store().Users().GetHubMe(ctx, userID)
	if err != nil {
		return nil, model.NewInternalError("App.GetHubCurrentUser", "failed to load hub user", err)
	}
	return user, nil
}

// GetInboxFolders يعيد مجلدات صندوق الوارد للمستخدم (مطابق لـ request8.txt).
func (a *YouTrackApp) GetInboxFolders(c request.CTX, userID string) ([]*model.InboxFolder, *model.AppError) {
	if userID == "" {
		userID = c.UserID()
	}
	ctx := c.Context()
	folders, err := a.Store().Users().InboxFolders(ctx, userID)
	if err != nil {
		return nil, model.NewInternalError("App.GetInboxFolders", "failed to load inbox folders", err)
	}
	return folders, nil
}

// GetInboxThreads يعيد خيوط الرسائل للمستخدم مع دعم الجلب الانتقائي (مطابق لـ request9.txt).
func (a *YouTrackApp) GetInboxThreads(c request.CTX, userID string, top int, skip int, tree *fields.FieldTree) ([]*model.InboxThread, *model.AppError) {
	if userID == "" {
		userID = c.UserID()
	}
	ctx := c.Context()
	threads, err := a.Store().Inbox().Threads(ctx, userID, top, skip, tree)
	if err != nil {
		return nil, model.NewInternalError("App.GetInboxThreads", "failed to load inbox threads", err)
	}
	return threads, nil
}
