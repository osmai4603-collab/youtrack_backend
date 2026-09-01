package app

import (
	"youtrack_backend/channels/model"
	"youtrack_backend/channels/model/fields"
	"youtrack_backend/channels/request"
)

// GetServices يعيد صفحة خدمات Hub مع الجلب الانتقائي حسب شجرة الحقول.
func (a *YouTrackApp) GetServices(c request.CTX, tree *fields.FieldTree, top int, skip int) (*model.ServicesPage, *model.AppError) {
	ctx := c.Context()
	page, err := a.Store().Admin().Services(ctx, tree, top, skip)
	if err != nil {
		return nil, model.NewInternalError("App.GetServices", "failed to load services", err)
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
