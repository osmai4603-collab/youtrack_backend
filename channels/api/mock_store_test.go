package api

import (
	"context"

	"youtrack_backend/channels/app"
	"youtrack_backend/channels/app/platform"
	"youtrack_backend/channels/store"
)

// mockStoreBase يكمل واجهة store.Store: المستودعات التسعة تعيد بقيمة صفرية
// ودوال البنية التحتية تعيد قيماً افتراضية آمنة.
// تُضمَّن في أي FullStore/حامل مخزن حتى يحقّق الواجهة كاملة (مؤشر *X يملك الواجهة).
type mockStoreBase struct{}

func (mockStoreBase) Users() store.UserStore                    { return nil }
func (mockStoreBase) Projects() store.ProjectStore              { return nil }
func (mockStoreBase) Issues() store.IssueStore                  { return nil }
func (mockStoreBase) Admin() store.AdminStore                   { return nil }
func (mockStoreBase) Inbox() store.InboxStore                   { return nil }
func (mockStoreBase) SavedQueries() store.SavedQueryStore       { return nil }
func (mockStoreBase) Search() store.SearchStore                 { return nil }
func (mockStoreBase) Subscriptions() store.SubscriptionStore    { return nil }
func (mockStoreBase) SecuritySearch() store.SecuritySearchStore { return nil }
func (mockStoreBase) Ready(ctx context.Context) error           { return nil }
func (mockStoreBase) Close()                                    {}
func (mockStoreBase) GetDbVersion() (string, error) {
	return "50", nil
}
func (mockStoreBase) GetDiagnostics(ctx context.Context) (map[string]any, error) {
	return nil, nil
}
func (mockStoreBase) TotalDbConnections() int { return 1 }

// newTestApp يبني كائن App مربوطاً بخادم مع مخزن وهمي محدد،
// بنفس نمط الإنتاج تماماً: PlatformService ← Server ← Channels ← App.
func newTestApp(st store.Store) *app.YouTrackApp {
	ps, err := platform.New(
		platform.ServiceOptionStore(st),
		platform.ServiceOptionJWTSecret("test-secret"),
	)
	if err != nil {
		panic(err)
	}
	server, err := app.NewServerWithOptions(ps)
	if err != nil {
		panic(err)
	}
	return app.New(app.ServerConnector(server.Channels()))
}
