package model

// InboxThread يمثّل خيط رسائل في صندوق الوارد.
type InboxThread struct {
	ID         string          `json:"id,omitempty"`
	Read       bool            `json:"read"`
	Muted      bool            `json:"muted"`
	Notified   bool            `json:"notified"`
	TargetType string          `json:"targetType,omitempty"`
	ThreadID   string          `json:"threadId,omitempty"`
	Timestamp  int64           `json:"timestamp,omitempty"`
	Updated    int64           `json:"updated,omitempty"`
	Messages   []*InboxMessage `json:"messages,omitempty"`
	Subject    *InboxSubject   `json:"subject,omitempty"`
	Type       string          `json:"$type"`
}

// InboxSubject يمثّل موضوع خيط الرسائل.
type InboxSubject struct {
	ID     string       `json:"id,omitempty"`
	Text   string       `json:"text,omitempty"`
	Target *InboxTarget `json:"target,omitempty"`
	Type   string       `json:"$type"`
}

// InboxTarget يمثّل الهدف المتعلق بالموضوع (مشكلة أو مقال).
type InboxTarget struct {
	ID             string          `json:"id,omitempty"`
	Content        string          `json:"content,omitempty"`
	Summary        string          `json:"summary,omitempty"`
	IDReadable     string          `json:"idReadable,omitempty"`
	Created        int64           `json:"created,omitempty"`
	Reporter       *User           `json:"reporter,omitempty"`
	MentionedUsers []*User         `json:"mentionedUsers,omitempty"`
	Attachments    []*Attachment   `json:"attachments,omitempty"`
	Issue          *Issue          `json:"issue,omitempty"`
	Article        *Article        `json:"article,omitempty"`
	Type           string          `json:"$type"`
}

// InboxMessage يمثّل رسالة داخل الخيط.
type InboxMessage struct {
	ID             string                `json:"id,omitempty"`
	Author         *User                 `json:"author,omitempty"`
	AuthorGroup    *UserGroup            `json:"authorGroup,omitempty"`
	Timestamp      int64                 `json:"timestamp,omitempty"`
	Text           string                `json:"text,omitempty"`
	Type           string                `json:"type,omitempty"`
	Pseudo         bool                  `json:"pseudo"`
	EmptyFieldText string                `json:"emptyFieldText,omitempty"`
	Params         []*InboxMessageParam  `json:"params,omitempty"`
	Activities     []*InboxActivity      `json:"activities,omitempty"`
	TypeField      string                `json:"$type"`
}

// InboxMessageParam يمثّل معامل داخل الرسالة.
type InboxMessageParam struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Type  string `json:"$type"`
}

// InboxActivity يمثّل نشاطاً متعلقاً بالرسالة.
type InboxActivity struct {
	ID       string            `json:"id,omitempty"`
	Category *ActivityCategory `json:"category,omitempty"`
	Added    any               `json:"added,omitempty"`
	Removed  any               `json:"removed,omitempty"`
	Comment  *Comment          `json:"comment,omitempty"`
	Issue    *Issue            `json:"issue,omitempty"`
	Article  *Article          `json:"article,omitempty"`
	Type     string            `json:"$type"`
}

// ActivityCategory يمثّل فئة النشاط.
type ActivityCategory struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
	Type string `json:"$type"`
}
