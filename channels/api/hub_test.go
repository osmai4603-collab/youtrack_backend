package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"

	"youtrack_backend/channels/model"
)

// sampleServices يبني صفحة خدمات مشابهة لـ request24.txt.
func sampleServices() *model.ServicesPage {
	trusted := true
	untrusted := false
	return &model.ServicesPage{
		Type:  "ServicesPage",
		Skip:  0,
		Top:   100,
		Total: 4,
		Services: []*model.HubService{
			{
				Type:            "service",
				ID:              "0-0-0-0-0",
				Name:            "YouTrack Administration",
				Key:             "jetbrains-hub-service",
				HomeURL:         "https://osm.youtrack.cloud/hub",
				ApplicationName: "Hub",
				Vendor:          "JetBrains",
				Version:         "2026.2.52366",
				Trusted:         &trusted,
			},
			{
				Type:            "service",
				ID:              "3fabcba6-15ff-405c-a050-ca52717d12eb",
				Name:            "YouTrack",
				Key:             "3fabcba6-15ff-405c-a050-ca52717d12eb",
				HomeURL:         "https://osm.youtrack.cloud",
				ApplicationName: "YouTrack",
				Vendor:          "JetBrains",
				Version:         "2026.2.18194",
				Trusted:         &trusted,
			},
			{
				Type:            "service",
				ID:              "db648fea-ba29-4757-899c-fe27c34e9473",
				Name:            "Konnector",
				Key:             "bad65d9b-3588-4e76-8f12-6146a85db8c7",
				HomeURL:         "https://konnector.services.jetbrains.com",
				ApplicationName: "Konnector",
				Vendor:          "JetBrains",
				Version:         "2026.2.18194",
				Trusted:         &trusted,
			},
			{
				Type:            "service",
				ID:              "92974ff4-3089-4a94-a38b-f9f3b1ca66f8",
				Name:            "YouTrack Mobile",
				Key:             "a715473a-8148-499e-8fb4-ac21fb464003",
				ApplicationName: "YouTrack Mobile",
				Vendor:          "JetBrains",
				Version:         "2026.2.18194",
				Trusted:         &untrusted,
			},
		},
	}
}

// TestServicesFullRequestFields يتحقق من المطابقة الدقيقة لاستجابة request24.txt:
// GET /hub/api/rest/services?fields=id,key,name,homeUrl,applicationName,vendor,version,trusted
func TestServicesFullRequestFields(t *testing.T) {
	a := newTestApp(&mockAdminStoreHolder{admin: &mockAdminStore{services: sampleServices()}})
	h := NewHubHandler(a)

	req := httptest.NewRequest("GET", "/hub/api/rest/services?fields=id,key,name,homeUrl,applicationName,vendor,version,trusted", nil)
	req = req.WithContext(withUserID(req.Context(), "2-1"))
	rec := httptest.NewRecorder()

	h.GetServices(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var page map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// المستوى الأول: type, skip, top, total, services
	if page["type"] != "ServicesPage" {
		t.Errorf("expected type ServicesPage, got %v", page["type"])
	}
	if page["skip"] != float64(0) {
		t.Errorf("expected skip 0, got %v", page["skip"])
	}
	if page["top"] != float64(100) {
		t.Errorf("expected top 100, got %v", page["top"])
	}
	if page["total"] != float64(4) {
		t.Errorf("expected total 4, got %v", page["total"])
	}

	services, ok := page["services"].([]any)
	if !ok || len(services) != 4 {
		t.Fatalf("expected 4 services, got %v", page["services"])
	}

	wantKeys := []string{"type", "id", "name", "key", "homeUrl", "applicationName", "vendor", "version", "trusted"}
	first := services[0].(map[string]any)
	if got := keysOf(first); !equalKeys(got, wantKeys) {
		t.Errorf("service keys mismatch: got %v want %v", got, wantKeys)
	}
	if first["type"] != "service" {
		t.Errorf("expected service type=service, got %v", first["type"])
	}
	if first["id"] != "0-0-0-0-0" {
		t.Errorf("expected id 0-0-0-0-0, got %v", first["id"])
	}
	if first["key"] != "jetbrains-hub-service" {
		t.Errorf("expected key jetbrains-hub-service, got %v", first["key"])
	}
	if first["homeUrl"] != "https://osm.youtrack.cloud/hub" {
		t.Errorf("unexpected homeUrl %v", first["homeUrl"])
	}
	if first["applicationName"] != "Hub" {
		t.Errorf("unexpected applicationName %v", first["applicationName"])
	}
	if first["vendor"] != "JetBrains" {
		t.Errorf("unexpected vendor %v", first["vendor"])
	}
	if first["version"] != "2026.2.52366" {
		t.Errorf("unexpected version %v", first["version"])
	}
	if first["trusted"] != true {
		t.Errorf("expected trusted true, got %v", first["trusted"])
	}

	last := services[3].(map[string]any)
	if last["trusted"] != false {
		t.Errorf("expected trusted false for YouTrack Mobile, got %v", last["trusted"])
	}
}

// TestServicesFieldSubset يتحقق من أن معامل fields يحدد الحقول الدقيقة فقط
// مع بقاء type ثابتاً (مثل طلبات hh.json).
func TestServicesFieldSubset(t *testing.T) {
	a := newTestApp(&mockAdminStoreHolder{admin: &mockAdminStore{services: sampleServices()}})
	h := NewHubHandler(a)

	req := httptest.NewRequest("GET", "/hub/api/rest/services?$top=-1&fields=applicationName,homeUrl,iconUrl,id,name,userUriPattern", nil)
	req = req.WithContext(withUserID(req.Context(), "2-1"))
	rec := httptest.NewRecorder()

	h.GetServices(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var page map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if page["top"] != float64(-1) {
		t.Errorf("expected top -1, got %v", page["top"])
	}

	services := page["services"].([]any)
	first := services[0].(map[string]any)
	wantKeys := []string{"type", "applicationName", "homeUrl", "iconUrl", "id", "name", "userUriPattern"}
	if got := keysOf(first); !equalKeys(got, wantKeys) {
		t.Errorf("subset keys mismatch: got %v want %v", got, wantKeys)
	}
	if _, exists := first["key"]; exists {
		t.Errorf("field key should NOT appear in subset response")
	}
	if _, exists := first["vendor"]; exists {
		t.Errorf("field vendor should NOT appear in subset response")
	}
	if _, exists := first["trusted"]; exists {
		t.Errorf("field trusted should NOT appear in subset response")
	}
	// iconUrl غير موجودة في البيانات فلا تُسلسَل كقيمة فارغة/null
	if v, ok := first["iconUrl"]; !ok || v != nil {
		t.Errorf("expected iconUrl null, got %v", first["iconUrl"])
	}
	if first["name"] != "YouTrack Administration" {
		t.Errorf("unexpected name %v", first["name"])
	}
}

// TestServicesOnlyID يتحقق من طلب fields=id&$top=1 (مطابق لـ hh.json السطر 36345).
func TestServicesOnlyID(t *testing.T) {
	a := newTestApp(&mockAdminStoreHolder{admin: &mockAdminStore{services: sampleServices()}})
	h := NewHubHandler(a)

	req := httptest.NewRequest("GET", "/hub/api/rest/services?fields=id&$top=1", nil)
	req = req.WithContext(withUserID(req.Context(), "2-1"))
	rec := httptest.NewRecorder()

	h.GetServices(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var page map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	services := page["services"].([]any)
	if len(services) != 1 {
		t.Fatalf("expected 1 service with $top=1, got %d", len(services))
	}
	first := services[0].(map[string]any)
	wantKeys := []string{"type", "id"}
	if got := keysOf(first); !equalKeys(got, wantKeys) {
		t.Errorf("only-id keys mismatch: got %v want %v", got, wantKeys)
	}
	if first["id"] != "0-0-0-0-0" {
		t.Errorf("expected id 0-0-0-0-0, got %v", first["id"])
	}
}

// TestServicesPagination يتحقق من $skip مع $top (التصفح).
func TestServicesPagination(t *testing.T) {
	a := newTestApp(&mockAdminStoreHolder{admin: &mockAdminStore{services: sampleServices()}})
	h := NewHubHandler(a)

	req := httptest.NewRequest("GET", "/hub/api/rest/services?fields=id&$skip=1&$top=1", nil)
	req = req.WithContext(withUserID(req.Context(), "2-1"))
	rec := httptest.NewRecorder()

	h.GetServices(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var page map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if page["skip"] != float64(1) {
		t.Errorf("expected skip 1, got %v", page["skip"])
	}
	if page["top"] != float64(1) {
		t.Errorf("expected top 1, got %v", page["top"])
	}
	services := page["services"].([]any)
	if len(services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(services))
	}
	id := services[0].(map[string]any)["id"]
	if id != "3fabcba6-15ff-405c-a050-ca52717d12eb" {
		t.Errorf("expected second service id after skip=1, got %v", id)
	}
}

// TestServicesDefaultTop يتحقق من أن الافتراضي لـ $top هو 100 (كما في request24.txt).
func TestServicesDefaultTop(t *testing.T) {
	a := newTestApp(&mockAdminStoreHolder{admin: &mockAdminStore{services: sampleServices()}})
	h := NewHubHandler(a)

	req := httptest.NewRequest("GET", "/hub/api/rest/services?fields=id", nil)
	req = req.WithContext(withUserID(req.Context(), "2-1"))
	rec := httptest.NewRecorder()

	h.GetServices(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	var page map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if page["top"] != float64(100) {
		t.Errorf("expected default top 100, got %v", page["top"])
	}
}

// TestServicesWithRouter يتحقق من تسجيل المسار عبر الراوتر وطبقة المصادقة JWT.
func TestServicesWithRouter(t *testing.T) {
	a := newTestApp(&mockAdminStoreHolder{admin: &mockAdminStore{services: sampleServices()}})
	router := NewRouter(a, "test-secret")

	req := httptest.NewRequest("GET", "/hub/api/rest/services?fields=id,name", nil)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   "2-1",
		"roles": []string{"user"},
	})
	tokenString, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var page map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if page["type"] != "ServicesPage" {
		t.Errorf("expected ServicesPage, got %v", page["type"])
	}
	for _, forbidden := range []string{"key", "homeUrl", "vendor"} {
		if _, exists := page["services"].([]any)[0].(map[string]any)[forbidden]; exists {
			t.Errorf("field %q should NOT appear", forbidden)
		}
	}
}

// TestServicesUnauthorized يتحقق من رفض الوصول دون توكن.
func TestServicesUnauthorized(t *testing.T) {
	a := newTestApp(&mockAdminStoreHolder{admin: &mockAdminStore{services: sampleServices()}})
	router := NewRouter(a, "test-secret")

	req := httptest.NewRequest("GET", "/hub/api/rest/services", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}
