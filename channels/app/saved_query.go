package app

import (
	"youtrack_backend/channels/model"
	"youtrack_backend/channels/model/fields"
	"youtrack_backend/channels/request"
)

// GetSavedQueries يعيد الاستعلامات المحفوظة مع الجلب الانتقائي.
func (a *YouTrackApp) GetSavedQueries(c request.CTX, tree *fields.FieldTree, top int, skip int) ([]*model.SavedQuery, *model.AppError) {
	ctx := c.Context()
	items, err := a.Store().SavedQueries().SavedQueries(ctx, tree, top, skip)
	if err != nil {
		return nil, model.NewInternalError("App.GetSavedQueries", "failed to load saved queries", err)
	}
	return items, nil
}
