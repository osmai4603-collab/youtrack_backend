package app

import (
	"context"

	"youtrack_backend/internal/api/fields"
	"youtrack_backend/internal/model"
)

// GetServices يعيد صفحة خدمات Hub مع الجلب الانتقائي حسب شجرة الحقول
// والحدود $top و $skip (مطابق لـ request24.txt و hh.json و hh2.json).
func (a *App) GetServices(ctx context.Context, tree *fields.FieldTree, top int, skip int) (*model.ServicesPage, error) {
	page, err := a.store.Admin().Services(ctx, tree, top, skip)
	if err != nil {
		return nil, model.Internal("failed to load services: %v", err)
	}
	if page == nil {
		page = &model.ServicesPage{Type: "ServicesPage", Services: []*model.HubService{}}
	}
	for _, s := range page.Services {
		if s != nil {
			s.Normalize()
		}
	}
	return page, nil
}
