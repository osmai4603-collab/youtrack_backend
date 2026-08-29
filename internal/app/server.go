package app

import (
	"youtrack_backend/internal/store"
)

// App يمثّل طبقة منطق الأعمال، ويجمّع المستودعات.
type App struct {
	store store.Store
}

// New ينشئ App من المستودعات المتاحة.
func New(s store.Store) *App {
	return &App{store: s}
}

// Store يعيد الوصول إلى المستودعات (للاستخدامات المتقدمة).
func (a *App) Store() store.Store { return a.store }
