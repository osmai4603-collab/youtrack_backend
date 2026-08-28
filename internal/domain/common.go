package domain

// Color تعريف الألوان المخصصة (للحقول والوسوم)
type Color struct {
	ID         string `json:"id"`
	Background string `json:"background,omitempty"`
	Foreground string `json:"foreground,omitempty"`
	Type       string `json:"$type,omitempty"`
}

// FieldStyle يمثل تنسيق الألوان الخاص بالحقل
type FieldStyle = Color

// Tag يمثل الوسم المرتبط بالتذكرة أو المقال
type Tag struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Color       *Color `json:"color,omitempty"`
	IsDeletable bool   `json:"isDeletable,omitempty"`
	IsUpdatable bool   `json:"isUpdatable,omitempty"`
	IsUsable    bool   `json:"isUsable,omitempty"`
	Type        string `json:"$type,omitempty"`
}

// Visibility إعدادات الخصوصية والظهور
type Visibility struct {
	PermittedUsers         []*User      `json:"permittedUsers,omitempty"`
	ImplicitPermittedUsers []*User      `json:"implicitPermittedUsers,omitempty"`
	PermittedGroups        []*UserGroup `json:"permittedGroups,omitempty"`
	Type                   string       `json:"$type,omitempty"`
}

// ImageDimensions أبعاد الصور
type ImageDimensions struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// Attachment يمثل المرفقات والملفات
type Attachment struct {
	ID              string           `json:"id"`
	Name            string           `json:"name"`
	Author          *User            `json:"author,omitempty"`
	Created         int64            `json:"created,omitempty"`
	Updated         int64            `json:"updated,omitempty"`
	MimeType        string           `json:"mimeType,omitempty"`
	URL             string           `json:"url,omitempty"`
	Size            int64            `json:"size,omitempty"`
	Visibility      *Visibility      `json:"visibility,omitempty"`
	ImageDimensions *ImageDimensions `json:"imageDimensions,omitempty"`
	ThumbnailURL    string           `json:"thumbnailURL,omitempty"`
	RecognizedText  string           `json:"recognizedText,omitempty"`
	Type            string           `json:"$type,omitempty"`
}
