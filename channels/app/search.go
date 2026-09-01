package app

import (
	"youtrack_backend/channels/model"
	"youtrack_backend/channels/model/fields"
	"youtrack_backend/channels/request"
)

// GetSearchAssist يخدم طلبات مساعد البحث.
func (a *YouTrackApp) GetSearchAssist(c request.CTX, query string, caret int, tree *fields.FieldTree) (*model.SearchAssistResponse, *model.AppError) {
	ctx := c.Context()
	resp, err := a.Store().Search().GetAssist(ctx, query, caret, tree)
	if err != nil {
		return nil, model.NewInternalError("App.GetSearchAssist", "failed to get search assist", err)
	}
	return resp, nil
}

// GetSecurityFilterFields يجلب حقول تصفية الأمان لنوع كيان معيّن.
func (a *YouTrackApp) GetSecurityFilterFields(c request.CTX, entityType string) ([]*model.SecurityFilterField, *model.AppError) {
	ctx := c.Context()
	resp, err := a.Store().SecuritySearch().GetFilterFields(ctx, entityType)
	if err != nil {
		return nil, model.NewInternalError("App.GetSecurityFilterFields", "failed to get security filter fields", err)
	}
	return resp, nil
}
