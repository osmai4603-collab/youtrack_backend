package app

import (
	"context"

	"youtrack_backend/internal/api/fields"
	"youtrack_backend/internal/model"
)

// GetSavedQueries يعيد الاستعلامات المحفوظة مع الجلب الانتقائي حسب شجرة
// الحقول (مطابق لـ request15.txt).
func (a *App) GetSavedQueries(ctx context.Context, tree *fields.FieldTree, top int, skip int) ([]*model.SavedQuery, error) {
	items, err := a.store.SavedQueries().SavedQueries(ctx, tree, top, skip)
	if err != nil {
		return nil, model.Internal("failed to load saved queries: %v", err)
	}
	return items, nil
}
