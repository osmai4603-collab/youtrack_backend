package domain

// NotificationTemplate قالب إشعار (مثل ملفات FreeMarker templates ftl)
type NotificationTemplate struct {
	ID          string `json:"id"`
	FileName    string `json:"fileName,omitempty"`
	Description string `json:"description,omitempty"`
	Content     string `json:"content,omitempty"`
	Overrided   bool   `json:"overrided,omitempty"`
	Type        string `json:"$type,omitempty"`
}

// NotificationTemplateGroup مجموعة تضم عدة قوالب إشعارات
type NotificationTemplateGroup struct {
	ID           string                       `json:"id"`
	Name         string                       `json:"name,omitempty"`
	Title        string                       `json:"title,omitempty"`
	Description  string                       `json:"description,omitempty"`
	Enabled      bool                         `json:"enabled,omitempty"`
	Customizable bool                         `json:"customizable,omitempty"`
	Subgroups    []*NotificationTemplateGroup `json:"subgroups,omitempty"`
	Templates    []*NotificationTemplate      `json:"templates,omitempty"`
	Type         string                       `json:"$type,omitempty"`
}

// InboxFolder مجلد الإشعارات الخاص بالمستخدم
type InboxFolder struct {
	ID           string `json:"id"`
	LastSeen     int64  `json:"lastSeen,omitempty"`
	LastNotified int64  `json:"lastNotified,omitempty"`
	Enabled      bool   `json:"enabled,omitempty"`
	Type         string `json:"$type,omitempty"`
}

// InboxThread خيط الإشعارات
type InboxThread struct {
	ID       string          `json:"id"`
	Subject  *InboxSubject   `json:"subject,omitempty"`
	Read     bool            `json:"read,omitempty"`
	Muted    bool            `json:"muted,omitempty"`
	Notified bool            `json:"notified,omitempty"`
	Messages []*InboxMessage `json:"messages,omitempty"`
	Type     string          `json:"$type,omitempty"`
}

// InboxSubject موضوع الإشعار
type InboxSubject struct {
	ID              string             `json:"id"`
	Target          any                `json:"target,omitempty"`
	Content         string             `json:"content,omitempty"`
	MentionedUsers  []*User            `json:"mentionedUsers,omitempty"`
	MentionedIssues []*Issue           `json:"mentionedIssues,omitempty"`
	MentionedArticles []*Article       `json:"mentionedArticles,omitempty"`
	Attachments     []*Attachment      `json:"attachments,omitempty"`
	Created         int64              `json:"created,omitempty"`
	Reporter        *InboxReporter     `json:"reporter,omitempty"`
	IDReadable      string             `json:"idReadable,omitempty"`
	Summary         string             `json:"summary,omitempty"`
	Text            string             `json:"text,omitempty"`
	Reactions       []*Reaction        `json:"reactions,omitempty"`
	ReactionOrder   string             `json:"reactionOrder,omitempty"`
	Visibility      *Visibility        `json:"visibility,omitempty"`
	Issue           *InboxSubjectIssue `json:"issue,omitempty"`
	Article         *InboxSubjectArticle `json:"article,omitempty"`
	Type            string             `json:"$type,omitempty"`
}

// InboxReporter يمثل مُبلّغ الإشعار
type InboxReporter struct {
	ID              string       `json:"id"`
	Login           string       `json:"login,omitempty"`
	Email           string       `json:"email,omitempty"`
	FullName        string       `json:"fullName,omitempty"`
	AvatarURL       string       `json:"avatarUrl,omitempty"`
	UserType        *UserType    `json:"userType,omitempty"`
	Name            string       `json:"name,omitempty"`
	IsEmailVerified bool         `json:"isEmailVerified,omitempty"`
	Guest           bool         `json:"guest,omitempty"`
	Online          bool         `json:"online,omitempty"`
	Banned          bool         `json:"banned,omitempty"`
	BanBadge        *string      `json:"banBadge,omitempty"`
	CanReadProfile  bool         `json:"canReadProfile,omitempty"`
	IsLocked        bool         `json:"isLocked,omitempty"`
	IssueRelatedGroup *UserGroup `json:"issueRelatedGroup,omitempty"`
	Profiles        *UserProfiles `json:"profiles,omitempty"`
	Type            string       `json:"$type,omitempty"`
}

// InboxSubjectIssue تفاصيل التذكرة في موضوع الإشعار
type InboxSubjectIssue struct {
	ID             string              `json:"id"`
	IDReadable     string              `json:"idReadable,omitempty"`
	Summary        string              `json:"summary,omitempty"`
	Description    string              `json:"description,omitempty"`
	Resolved       *int64              `json:"resolved,omitempty"`
	Created        int64               `json:"created,omitempty"`
	Updated        int64               `json:"updated,omitempty"`
	UnauthenticatedReporter bool       `json:"unauthenticatedReporter,omitempty"`
	Fields         []*IssueCustomField `json:"fields,omitempty"`
	Project        *Project            `json:"project,omitempty"`
	Creator        *User               `json:"creator,omitempty"`
	Reporter       *User               `json:"reporter,omitempty"`
	Visibility     *Visibility         `json:"visibility,omitempty"`
	Tags           []*Tag              `json:"tags,omitempty"`
	Votes          int                 `json:"votes,omitempty"`
	Voters         *IssueVoters        `json:"voters,omitempty"`
	Watchers       *IssueWatchers      `json:"watchers,omitempty"`
	UsersTyping    []*UserTyping       `json:"usersTyping,omitempty"`
	CanUndoComment bool                `json:"canUndoComment,omitempty"`
	CanAddPublicComment bool           `json:"canAddPublicComment,omitempty"`
	Type           string              `json:"$type,omitempty"`
}

// InboxSubjectArticle تفاصيل المقال في موضوع الإشعار
type InboxSubjectArticle struct {
	ID            string       `json:"id"`
	IDReadable    string       `json:"idReadable,omitempty"`
	Summary       string       `json:"summary,omitempty"`
	Project       *Project     `json:"project,omitempty"`
	ParentArticle *Article     `json:"parentArticle,omitempty"`
	Ordinal       int          `json:"ordinal,omitempty"`
	Visibility    *Visibility  `json:"visibility,omitempty"`
	HasStar       bool         `json:"hasStar,omitempty"`
	Updated       int64        `json:"updated,omitempty"`
	Type          string       `json:"$type,omitempty"`
}

// IssueVoters يمثل قائمة الم窝طئين
type IssueVoters struct {
	HasVote bool `json:"hasVote,omitempty"`
}

// IssueWatchers يمثل قائمة المراقبين
type IssueWatchers struct {
	HasStar bool `json:"hasStar,omitempty"`
}

// UserTyping يمثل المستخدم الذي يكتب حالياً
type UserTyping struct {
	Timestamp int64 `json:"timestamp,omitempty"`
	User      *User `json:"user,omitempty"`
}

// InboxMessage رسالة في خيط الإشعار
type InboxMessage struct {
	ID       string `json:"id"`
	Read     bool   `json:"read,omitempty"`
	Target   string `json:"targetType,omitempty"`
	ThreadID string `json:"threadId,omitempty"`
	Timestamp int64 `json:"timestamp,omitempty"`
	Text     string `json:"text,omitempty"`
	Author   *User  `json:"author,omitempty"`
	Type     string `json:"$type,omitempty"`
}

// NotificationSupplement معلومات تكميلية للإشعارات
type NotificationSupplement struct {
	Preview *NotificationPreview `json:"preview,omitempty"`
	Type    string               `json:"$type,omitempty"`
}

// NotificationPreview معاينة الإشعار
type NotificationPreview struct {
	IssueID string `json:"issueId,omitempty"`
	Type    string `json:"$type,omitempty"`
}
