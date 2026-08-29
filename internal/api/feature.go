package api

import (
	"net/http"

	"youtrack_backend/internal/api/fields"
	"youtrack_backend/internal/app"
	"youtrack_backend/internal/model"
)

// FeatureHandler يعالج طلبات ملف الميزات (request10.txt).
type FeatureHandler struct {
	app *app.App
}

func NewFeatureHandler(a *app.App) *FeatureHandler {
	return &FeatureHandler{app: a}
}

// Get يعيد ملف ميزات YouTrack (en_US) مباشرة في جذر الـ JSON، مطابقاً
// لـ https://resources.jetbrains.com/youtrack/features/features-en_US.json،
// مع احترام معامل fields للجلب الانتقائي على مستوى النسخ والميزات.
func (h *FeatureHandler) Get(w http.ResponseWriter, r *http.Request) {
	fieldTree := fields.Parse(r.URL.Query().Get("fields"))
	versions := model.Features()
	writeModel(w, featuresToMap(versions, fieldTree))
}

// featuresToMap يحوّل بيانات الميزات إلى خريطة بحقل "versions" في الجذر،
// مع احترام معامل fields وفق بنية versions -> features(header, content, doc).
func featuresToMap(versions []model.FeatureVersion, tree *fields.FieldTree) map[string]any {
	result := make(map[string]any)

	if tree == nil || tree.IsEmpty() || tree.Has("versions") {
		result["versions"] = featureVersionsToAny(versions, tree)
	}
	return result
}

// featureVersionsToAny يحوّل نسخ الإصدارات إلى مصفوفة خريطة تحترم شجرة الحقول.
func featureVersionsToAny(versions []model.FeatureVersion, tree *fields.FieldTree) []map[string]any {
	versionsTree := tree
	if tree != nil {
		if child := tree.Child("versions"); child != nil {
			versionsTree = child
		}
	}

	out := make([]map[string]any, 0, len(versions))
	for _, v := range versions {
		item := make(map[string]any)
		if versionsTree == nil || versionsTree.IsEmpty() || versionsTree.Has("id") {
			item["id"] = v.ID
		}
		if versionsTree == nil || versionsTree.IsEmpty() || versionsTree.Has("features") {
			item["features"] = featuresToAny(v.Features, versionsTree.Child("features"))
		}
		out = append(out, item)
	}
	return out
}

// featuresToAny يحوّل قائمة الميزات إلى مصفوفة خريطة تحترم شجرة الحقول الفرعية.
func featuresToAny(features []model.Feature, tree *fields.FieldTree) []map[string]any {
	out := make([]map[string]any, 0, len(features))
	for _, f := range features {
		item := make(map[string]any)
		if tree == nil || tree.IsEmpty() || tree.Has("header") {
			item["header"] = f.Header
		}
		if tree == nil || tree.IsEmpty() || tree.Has("content") {
			item["content"] = f.Content
		}
		if tree == nil || tree.IsEmpty() || tree.Has("doc") {
			item["doc"] = f.Doc
		}
		out = append(out, item)
	}
	return out
}
