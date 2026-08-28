package fields

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantNil  bool
		wantKeys []string
	}{
		{
			name:    "فارغ يُرجع nil",
			input:   "",
			wantNil: true,
		},
		{
			name:     "حقول بسيطة",
			input:    "id,name,email",
			wantKeys: []string{"id", "name", "email"},
		},
		{
			name:     "حقل فرعي واحد",
			input:    "id,project(id,name)",
			wantKeys: []string{"id", "project"},
		},
		{
			name:     "حقول فرعية متداخلة",
			input:    "id,project(id,name,shortName,projectType(id))",
			wantKeys: []string{"id", "project"},
		},
		{
			name:     "حقول متعددة مع فرعية",
			input:    "id,login,email,fullName,avatarUrl,userType(id,name)",
			wantKeys: []string{"id", "login", "email", "fullName", "avatarUrl", "userType"},
		},
		{
			name:     "تداخل عميق",
			input:    "a(b(c(d),e),f)",
			wantKeys: []string{"a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := Parse(tt.input)

			if tt.wantNil {
				if tree != nil {
					t.Errorf("Parse(%q) expected nil, got %+v", tt.input, tree)
				}
				return
			}

			if tree == nil {
				t.Fatalf("Parse(%q) returned nil, expected non-nil", tt.input)
			}

			if len(tree.Children) != len(tt.wantKeys) {
				t.Errorf("Parse(%q) expected %d children, got %d", tt.input, len(tt.wantKeys), len(tree.Children))
			}

			for _, key := range tt.wantKeys {
				if _, exists := tree.Children[key]; !exists {
					t.Errorf("Parse(%q) missing child %q", tt.input, key)
				}
			}
		})
	}
}

func TestParseWithDeclaredAliases(t *testing.T) {
	// نفس شاكلة request1.txt: تعريفات @helpdeskContext و @permittedUsers و @permittedGroups
	// مفصولة بفاصلة منقوطة، مع إشارات إليها داخل شجرة الحقول.
	input := "id,login,issueRelatedGroup(@permittedGroups),profiles(" +
		"general(searchContext(@helpdeskContext),helpdeskContext(@helpdeskContext))," +
		"articles(lastVisitedArticle(reporter(@permittedUsers))))" +
		";@helpdeskContext:id,name,issuesUrl,pinned,pinnedInHelpdesk,owner(@permittedUsers),$type,query,isUpdatable,shortName" +
		";@permittedUsers:id,login,email,fullName,avatarUrl,userType(id,name)" +
		";@permittedGroups:id,name,$type(),auditTargetId,description,allUsersGroup,icon,teamForProject(id,name,icon),isUpdatable,isRemovable"

	tree := Parse(input)
	if tree == nil {
		t.Fatal("Parse returned nil")
	}

	// issueRelatedGroup يجب أن يتوسع إلى حقول @permittedGroups
	group, ok := tree.Children["issueRelatedGroup"]
	if !ok {
		t.Fatal("Missing 'issueRelatedGroup' child")
	}
	for _, key := range []string{"id", "name", "$type", "auditTargetId", "description",
		"allUsersGroup", "icon", "teamForProject", "isUpdatable", "isRemovable"} {
		if _, exists := group.Children[key]; !exists {
			t.Errorf("issueRelatedGroup missing expanded field %q", key)
		}
	}
	team, ok := group.Children["teamForProject"]
	if !ok || !team.HasChildren() {
		t.Fatal("teamForProject should have children (id,name,icon)")
	}
	for _, key := range []string{"id", "name", "icon"} {
		if _, exists := team.Children[key]; !exists {
			t.Errorf("teamForProject missing child %q", key)
		}
	}

	// لا يجب أن تظهر أي عقدة باسم "@permittedGroups" الحرفي
	if _, exists := group.Children["@permittedGroups"]; exists {
		t.Error("literal '@permittedGroups' node should not exist after expansion")
	}

	// searchContext داخل general يجب أن يتوسع إلى حقول @helpdeskContext
	general := tree.Children["profiles"].Children["general"]
	sc, ok := general.Children["searchContext"]
	if !ok || !sc.HasChildren() {
		t.Fatal("searchContext should have children from @helpdeskContext")
	}
	for _, key := range []string{"id", "name", "issuesUrl", "pinned", "pinnedInHelpdesk",
		"owner", "$type", "query", "isUpdatable", "shortName"} {
		if _, exists := sc.Children[key]; !exists {
			t.Errorf("searchContext missing expanded field %q", key)
		}
	}

	// owner داخل helpdeskContext يجب أن يتوسع إلى @permittedUsers المتداخل
	owner, ok := sc.Children["owner"]
	if !ok {
		t.Fatal("searchContext.owner missing")
	}
	for _, key := range []string{"id", "login", "email", "fullName", "avatarUrl", "userType"} {
		if _, exists := owner.Children[key]; !exists {
			t.Errorf("owner missing expanded field %q", key)
		}
	}

	// reporter داخل lastVisitedArticle يجب أن يتوسع إلى @permittedUsers
	article := tree.Children["profiles"].Children["articles"].Children["lastVisitedArticle"]
	reporter, ok := article.Children["reporter"]
	if !ok {
		t.Fatal("lastVisitedArticle.reporter missing")
	}
	for _, key := range []string{"id", "login", "email", "userType"} {
		if _, exists := reporter.Children[key]; !exists {
			t.Errorf("reporter missing expanded field %q", key)
		}
	}
}

func TestParseBuiltinAliases(t *testing.T) {
	// الأسماء المستعارة المدمجة تُوسَّع حتى دون تعريف صريح داخل المعامل
	tree := Parse("id,reporter(@permittedUsers)")
	if tree == nil {
		t.Fatal("Parse returned nil")
	}
	reporter := tree.Children["reporter"]
	if !reporter.HasChildren() {
		t.Fatal("reporter should have children from builtin @permittedUsers")
	}
	if _, exists := reporter.Children["isLocked"]; !exists {
		t.Error("builtin @permittedUsers should include isLocked")
	}
	if _, exists := reporter.Children["@permittedUsers"]; exists {
		t.Error("literal '@permittedUsers' node should not exist after expansion")
	}
}

func TestParseAliasUnknownKeptLiteral(t *testing.T) {
	// اسم مستعار غير معروف يُبقى كعقدة حرفية ولا يكسر التحليل
	tree := Parse("id,@unknownMacro")
	if tree == nil {
		t.Fatal("Parse returned nil")
	}
	if _, exists := tree.Children["@unknownMacro"]; !exists {
		t.Error("unknown macro should be kept as a literal node")
	}
	if _, exists := tree.Children["id"]; !exists {
		t.Error("id should remain present")
	}
}

func TestParseNested(t *testing.T) {
	tree := Parse("id,project(id,name,shortName,team(id,name,users(id,login,email)))")
	if tree == nil {
		t.Fatal("Parse returned nil")
	}

	// التحقق من project
	project, ok := tree.Children["project"]
	if !ok {
		t.Fatal("Missing 'project' child")
	}
	if !project.HasChildren() {
		t.Fatal("project should have children")
	}

	// التحقق من project.team.users
	team, ok := project.Children["team"]
	if !ok {
		t.Fatal("Missing 'team' in project")
	}
	users, ok := team.Children["users"]
	if !ok {
		t.Fatal("Missing 'users' in team")
	}
	if !users.HasChildren() {
		t.Fatal("users should have children")
	}
	if _, ok := users.Children["id"]; !ok {
		t.Fatal("Missing 'id' in users")
	}
	if _, ok := users.Children["login"]; !ok {
		t.Fatal("Missing 'login' in users")
	}
}

func TestFilterMap(t *testing.T) {
	data := map[string]any{
		"id":       "11-2095841",
		"login":    "osmflutterdeveloper",
		"email":    "osmflutterdeveloper@gmail.com",
		"fullName": "OSM ABASSI",
		"banned":   false,
		"online":   true,
		"userType": map[string]any{
			"id":    "STANDARD_USER",
			"name":  "Standard user",
			"$type": "UserType",
		},
		"project": map[string]any{
			"id":        "0-0",
			"name":      "Demo project",
			"shortName": "DEMO",
			"archived":  false,
		},
	}

	tree := Parse("id,login,fullName,userType(id,name)")
	result := Filter(data, tree).(map[string]any)

	// يجب أن يحتوي على الحقول المطلوبة فقط
	expectedKeys := []string{"id", "login", "fullName", "userType"}
	for _, key := range expectedKeys {
		if _, exists := result[key]; !exists {
			t.Errorf("Missing key %q in filtered result", key)
		}
	}

	// يجب أن يحتوي على حقول userType الفرعية فقط
	userType := result["userType"].(map[string]any)
	if _, exists := userType["id"]; !exists {
		t.Error("Missing 'id' in userType")
	}
	if _, exists := userType["name"]; !exists {
		t.Error("Missing 'name' in userType")
	}
	// $type يُحفظ دائماً حتى لو لم يُطلب (سلوك YouTrack الموثّق في request1.txt)
	if _, exists := userType["$type"]; !exists {
		t.Error("$type should always be preserved in userType")
	}

	// يجب أن لا يحتوي على banned و online و project
	for _, key := range []string{"banned", "online", "email", "project"} {
		if _, exists := result[key]; exists {
			t.Errorf("Unexpected key %q in filtered result", key)
		}
	}
}

func TestFilterNilTree(t *testing.T) {
	data := map[string]any{
		"id":   "11-2095841",
		"name": "Test",
	}

	result := Filter(data, nil)
	resultMap := result.(map[string]any)

	if !reflect.DeepEqual(data, resultMap) {
		t.Errorf("Filter with nil tree should return data as-is")
	}
}

func TestFilterSlice(t *testing.T) {
	data := []any{
		map[string]any{"id": "1", "name": "Test1", "extra": "a"},
		map[string]any{"id": "2", "name": "Test2", "extra": "b"},
	}

	tree := Parse("id,name")
	result := Filter(data, tree).([]any)

	if len(result) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(result))
	}

	for i, item := range result {
		m := item.(map[string]any)
		if _, exists := m["id"]; !exists {
			t.Errorf("Item %d: missing 'id'", i)
		}
		if _, exists := m["name"]; !exists {
			t.Errorf("Item %d: missing 'name'", i)
		}
		if _, exists := m["extra"]; exists {
			t.Errorf("Item %d: 'extra' should be filtered out", i)
		}
	}
}

func TestFilterNestedObject(t *testing.T) {
	data := map[string]any{
		"id": "11-2095841",
		"project": map[string]any{
			"id":        "0-0",
			"name":      "Demo",
			"shortName": "DEMO",
			"archived":  false,
		},
	}

	tree := Parse("id,project(id,name)")
	result := Filter(data, tree).(map[string]any)

	project := result["project"].(map[string]any)
	if len(project) != 2 {
		t.Errorf("project should have 2 fields, got %d: %v", len(project), project)
	}
	if project["id"] != "0-0" {
		t.Errorf("project.id = %v, want '0-0'", project["id"])
	}
	if project["name"] != "Demo" {
		t.Errorf("project.name = %v, want 'Demo'", project["name"])
	}
}

func TestFilterDeeplyNested(t *testing.T) {
	data := map[string]any{
		"id": "11-2095841",
		"userType": map[string]any{
			"id":   "STANDARD_USER",
			"name": "Standard user",
		},
		"project": map[string]any{
			"id":   "0-0",
			"name": "Demo",
			"projectType": map[string]any{
				"id": "DEFAULT",
			},
		},
	}

	tree := Parse("id,userType(id),project(id,projectType(id))")
	result := Filter(data, tree).(map[string]any)

	// userType: فقط id
	userType := result["userType"].(map[string]any)
	if len(userType) != 1 {
		t.Errorf("userType should have 1 field, got %d", len(userType))
	}

	// project: id + projectType(id)
	project := result["project"].(map[string]any)
	if len(project) != 2 {
		t.Errorf("project should have 2 fields (id, projectType), got %d", len(project))
	}
	pt := project["projectType"].(map[string]any)
	if len(pt) != 1 {
		t.Errorf("projectType should have 1 field, got %d", len(pt))
	}
}

func TestHasField(t *testing.T) {
	tree := Parse("id,name,project(id,name)")

	if !HasField(tree, "id") {
		t.Error("HasField(id) should be true")
	}
	if !HasField(tree, "name") {
		t.Error("HasField(name) should be true")
	}
	if HasField(tree, "email") {
		t.Error("HasField(email) should be false")
	}
	if !HasField(nil, "anything") {
		t.Error("HasField(nil, anything) should be true (no filter)")
	}
}

func TestGetChild(t *testing.T) {
	tree := Parse("id,project(id,name)")

	child := GetChild(tree, "project")
	if child == nil {
		t.Fatal("GetChild(project) returned nil")
	}
	if !child.HasChildren() {
		t.Error("project child should have children")
	}

	if GetChild(tree, "nonexistent") != nil {
		t.Error("GetChild(nonexistent) should be nil")
	}
}

func TestToJSON(t *testing.T) {
	tree := Parse("id,name,project(id,name)")

	result := ToJSON(tree)
	if result == nil {
		t.Fatal("ToJSON returned nil")
	}

	if result["id"] != true {
		t.Error("id should be true")
	}
	if result["name"] != true {
		t.Error("name should be true")
	}

	// project should be a nested map
	projectMap, ok := result["project"].(map[string]any)
	if !ok {
		t.Fatalf("project should be a map, got %T", result["project"])
	}
	if projectMap["id"] != true {
		t.Error("project.id should be true")
	}
}

func TestFilterEmptyMap(t *testing.T) {
	data := map[string]any{}
	tree := Parse("id,name")
	result := Filter(data, tree).(map[string]any)

	if len(result) != 0 {
		t.Errorf("Expected empty result, got %d items", len(result))
	}
}

func TestFilterMissingFields(t *testing.T) {
	data := map[string]any{
		"id":   "11-2095841",
		"name": "Test",
	}

	tree := Parse("id,name,nonexistent")
	result := Filter(data, tree).(map[string]any)

	// الحقول المفقودة يتم تخطيها
	if len(result) != 2 {
		t.Errorf("Expected 2 items (missing fields skipped), got %d", len(result))
	}
}
