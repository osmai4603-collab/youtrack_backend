package app

import (
	"context"

	"youtrack_backend/internal/api/fields"
	"youtrack_backend/internal/model"
)

// GetSearchAssist يخدم طلبات مساعد البحث.
func (a *App) GetSearchAssist(ctx context.Context, query string, caret int, tree *fields.FieldTree) (*model.SearchAssistResponse, error) {
	return a.store.Search().GetAssist(ctx, query, caret, tree)
}
