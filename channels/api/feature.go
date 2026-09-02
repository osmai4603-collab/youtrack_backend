package api

import (
	"net/http"

	"youtrack_backend/channels/app"
	"youtrack_backend/channels/model"
	"youtrack_backend/channels/model/fields"
)

// FeatureHandler يعالج طلبات ملف الميزات (request10.txt).
type FeatureHandler struct {
	app *app.YouTrackApp
}

func NewFeatureHandler(a *app.YouTrackApp) *FeatureHandler {
	return &FeatureHandler{app: a}
}

func (api *API) InitFeature() {
	handler := NewFeatureHandler(api.newApp())
	api.BaseRoutes.Root.Handle("/static/features-en_US.json", api.APIHandler(func(c *Context, w http.ResponseWriter, r *http.Request) {
		handler.Get(w, r)
	})).Methods("GET")
}

// Get يعيد ملف ميزات YouTrack (en_US).
func (h *FeatureHandler) Get(w http.ResponseWriter, r *http.Request) {
	fieldTree := fields.Parse(r.URL.Query().Get("fields"))
	versions := model.Features()
	writeModel(w, featuresToMap(versions, fieldTree))
}

func featuresToMap(versions []model.FeatureVersion, tree *fields.FieldTree) map[string]any {
	result := make(map[string]any)
	if tree == nil || tree.IsEmpty() || tree.Has("versions") {
		result["versions"] = featureVersionsToAny(versions, tree)
	}
	return result
}

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
