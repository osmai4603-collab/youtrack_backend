package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"youtrack_backend/channels/app"
)

func TestFeaturesEndpoint(t *testing.T) {
	h := NewFeatureHandler(app.New(nil))

	req := httptest.NewRequest(http.MethodGet, "/static/features-en_US.json", nil)
	rec := httptest.NewRecorder()
	h.Get(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var res struct {
		Versions []map[string]any `json:"versions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if len(res.Versions) == 0 {
		t.Fatal("expected non-empty versions")
	}

	// كل نسخة يجب أن تحمل id و features كاملة (بدون fields)
	first := res.Versions[0]
	if _, ok := first["id"]; !ok {
		t.Errorf("expected version to include 'id', got %v", first)
	}
	if _, ok := first["features"]; !ok {
		t.Errorf("expected version to include 'features', got %v", first)
	}
	feats, ok := first["features"].([]any)
	if !ok || len(feats) == 0 {
		t.Fatalf("expected non-empty features array, got %v", first["features"])
	}
	f0 := feats[0].(map[string]any)
	for _, k := range []string{"header", "content", "doc"} {
		if _, ok := f0[k]; !ok {
			t.Errorf("expected feature to include %s, got %v", k, f0)
		}
	}
}

func TestFeaturesEndpointFieldsFiltering(t *testing.T) {
	h := NewFeatureHandler(app.New(nil))

	req := httptest.NewRequest(http.MethodGet, "/static/features-en_US.json?fields=versions(id,features(header,doc))", nil)
	rec := httptest.NewRecorder()
	h.Get(rec, req)

	var res struct {
		Versions []map[string]any `json:"versions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	first := res.Versions[0]
	if _, ok := first["id"]; !ok {
		t.Errorf("expected version to include requested 'id', got %v", first)
	}
	feats, ok := first["features"].([]any)
	if !ok || len(feats) == 0 {
		t.Fatalf("expected features array, got %v", first["features"])
	}
	f0 := feats[0].(map[string]any)
	if _, ok := f0["header"]; !ok {
		t.Errorf("expected feature to include requested 'header', got %v", f0)
	}
	if _, ok := f0["doc"]; !ok {
		t.Errorf("expected feature to include requested 'doc', got %v", f0)
	}
	if _, ok := f0["content"]; ok {
		t.Errorf("did not expect 'content' (not requested), got %v", f0)
	}
}

func TestFeaturesEndpointNestedField(t *testing.T) {
	h := NewFeatureHandler(app.New(nil))

	// طلب حقل نسخ فقط (بدون features) -> إخفاء الميزات
	req := httptest.NewRequest(http.MethodGet, "/static/features-en_US.json?fields=versions(id)", nil)
	rec := httptest.NewRecorder()
	h.Get(rec, req)

	var res struct {
		Versions []map[string]any `json:"versions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if _, ok := res.Versions[0]["features"]; ok {
		t.Error("did not expect 'features' when only versions(id) requested")
	}
	if _, ok := res.Versions[0]["id"]; !ok {
		t.Error("expected 'id' to be present")
	}
}

func TestFeaturesRouteRegistered(t *testing.T) {
	a := app.New(nil)
	router := NewRouter(a, "test-secret")

	req := httptest.NewRequest(http.MethodGet, "/static/features-en_US.json", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected route GET /static/features-en_US.json to return 200 (unauthenticated), got %d", rec.Code)
	}
}
