package model

// ProjectDashboard يمثّل لوحة ودجات لوحة مشروع (مطابقة لـ request13.txt).
// مخطط جديد مستقل: لا يعدّل أي مخطط موجود.
type ProjectDashboard struct {
	Widgets []*ProjectDashboardWidget `json:"widgets,omitempty"`
	Type    string                    `json:"$type,omitempty"`
}

// Normalize يملأ $type للوحة المشروع.
func (d *ProjectDashboard) Normalize() {
	d.Type = "ProjectDashboard"
}

// ProjectDashboardWidget يمثّل ودجت واحد في لوحة مشروع (مطابقة لـ request13.txt).
type ProjectDashboardWidget struct {
	ID       string           `json:"id,omitempty"`
	Key      string           `json:"key,omitempty"`
	X        int              `json:"x,omitempty"`
	Y        int              `json:"y,omitempty"`
	Width    int              `json:"width,omitempty"`
	Height   int              `json:"height,omitempty"`
	Widget   *DashboardWidget `json:"widget,omitempty"`
	Settings string           `json:"settings,omitempty"`
	Type     string           `json:"$type,omitempty"`
}

// Normalize يملأ $type للودجت.
func (w *ProjectDashboardWidget) Normalize() {
	w.Type = "ProjectDashboardWidget"
	if w.Widget != nil && w.Widget.Type == "" {
		w.Widget.Type = "WidgetView"
	}
}