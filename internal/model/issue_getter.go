package model

// IssueGetterIssue يمثّل الاستجابة المخصصة لـ /api/issuesGetter (طلب #28).
// صُمم هذا المخطط ليكون مستقلاً عن Issue الأساسي لتجنب التعديلات الجانبية.
type IssueGetterIssue struct {
	ID         string              `json:"id"`
	IDReadable string              `json:"idReadable"`
	Summary    string              `json:"summary"`
	Resolved   *int64              `json:"resolved"`
	Fields     []*IssueCustomField `json:"fields"`
	Type       string              `json:"$type"`
}

func (i *IssueGetterIssue) Normalize() {
	i.Type = "Issue"
	for _, f := range i.Fields {
		f.Normalize()
	}
}

type IssueCustomField struct {
	ID                 string              `json:"id"`
	Name               string              `json:"name,omitempty"`
	Value              any                 `json:"value"` // يمكن أن يكون IssueFieldValue أو []*IssueFieldValue
	ProjectCustomField *ProjectCustomField `json:"projectCustomField"`
	Type               string              `json:"$type"`
}

func (f *IssueCustomField) Normalize() {
	f.Type = "IssueCustomField"
	if f.ProjectCustomField != nil {
		f.ProjectCustomField.Normalize()
	}
}

type IssueFieldValue struct {
	ID            string      `json:"id,omitempty"`
	Name          string      `json:"name,omitempty"`
	LocalizedName string      `json:"localizedName,omitempty"`
	Login         string      `json:"login,omitempty"`
	AvatarURL     string      `json:"avatarUrl,omitempty"`
	Presentation  string      `json:"presentation,omitempty"`
	Minutes       int         `json:"minutes,omitempty"`
	Color         *FieldColor `json:"color,omitempty"`
	Type          string      `json:"$type,omitempty"`
}

type FieldColor struct {
	ID         string `json:"id"`
	Foreground string `json:"foreground,omitempty"`
	Background string `json:"background,omitempty"`
	Type       string `json:"$type"`
}

type ProjectCustomField struct {
	ID     string               `json:"id"`
	Bundle *FieldBundle         `json:"bundle"`
	Field  *CustomFieldMetadata `json:"field"`
	Type   string               `json:"$type"`
}

func (pcf *ProjectCustomField) Normalize() {
	pcf.Type = "ProjectCustomField"
	if pcf.Bundle != nil && pcf.Bundle.Type == "" {
		pcf.Bundle.Type = "Bundle"
	}
	if pcf.Field != nil {
		pcf.Field.Normalize()
	}
}

type FieldBundle struct {
	ID   string `json:"id"`
	Type string `json:"$type"`
}

type CustomFieldMetadata struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	LocalizedName string     `json:"localizedName"`
	FieldType     *FieldType `json:"fieldType"`
	Type          string     `json:"$type"`
}

func (cfm *CustomFieldMetadata) Normalize() {
	cfm.Type = "CustomField"
	if cfm.FieldType != nil && cfm.FieldType.Type == "" {
		cfm.FieldType.Type = "FieldType"
	}
}

type FieldType struct {
	ID        string `json:"id"`
	ValueType string `json:"valueType"`
	Type      string `json:"$type"`
}
