package domain

// SearchAssist نتيجة مساعدة البحث
type SearchAssist struct {
	Caret                    int            `json:"caret"`
	Query                    string         `json:"query"`
	StyleRanges              []*StyleRange  `json:"styleRanges,omitempty"`
	Suggestions              []*Suggestion  `json:"suggestions,omitempty"`
	SortProperties           any            `json:"sortProperties,omitempty"`
	IgnoreUnresolvedSetting  bool           `json:"ignoreUnresolvedSetting,omitempty"`
	QueryFeatures            *QueryFeatures `json:"queryFeatures,omitempty"`
	AST                      any            `json:"ast,omitempty"`
	Type                     string         `json:"$type,omitempty"`
}

// StyleRange نطاق تنسيق في نص الاستعلام
type StyleRange struct {
	Start int    `json:"start"`
	Length int   `json:"length"`
	Style string `json:"style,omitempty"`
	Title string `json:"title,omitempty"`
}

// Suggestion اقتراح إكمال تلقائي للاستعلام
type Suggestion struct {
	Description    string `json:"description,omitempty"`
	Group          string `json:"group,omitempty"`
	Icon           string `json:"icon,omitempty"`
	Option         string `json:"option,omitempty"`
	Prefix         string `json:"prefix,omitempty"`
	Suffix         string `json:"suffix,omitempty"`
	ClassName      string `json:"className,omitempty"`
	MatchingStart  int    `json:"matchingStart,omitempty"`
	MatchingEnd    int    `json:"matchingEnd,omitempty"`
	Caret          int    `json:"caret,omitempty"`
	CompletionStart int   `json:"completionStart,omitempty"`
	CompletionEnd  int    `json:"completionEnd,omitempty"`
}

// QueryFeatures خصائص الاستعلام المُحلل
type QueryFeatures struct {
	ContainsWildcard           bool   `json:"containsWildcard,omitempty"`
	ContainsLatin              bool   `json:"containsLatin,omitempty"`
	ContainsDigits             bool   `json:"containsDigits,omitempty"`
	ContainsNotLatin           bool   `json:"containsNotLatin,omitempty"`
	NumberOfWords              int    `json:"numberOfWords,omitempty"`
	HasAnd                     bool   `json:"hasAnd,omitempty"`
	HasOr                      bool   `json:"hasOr,omitempty"`
	HasSorting                 bool   `json:"hasSorting,omitempty"`
	Length                     int    `json:"length,omitempty"`
	QueryLength                int    `json:"queryLength,omitempty"`
	FieldsCount                int    `json:"fieldsCount,omitempty"`
	UserInFields               bool   `json:"userInFields,omitempty"`
	MeInFields                 bool   `json:"meInFields,omitempty"`
	DatePeriodInFields         bool   `json:"datePeriodInFields,omitempty"`
	SortedByRelevance          bool   `json:"sortedByRelevance,omitempty"`
	SortByRelevanceSetting     bool   `json:"sortByRelevanceSetting,omitempty"`
	ExperimentVersion          string `json:"experimentVersion,omitempty"`
	ExperimentGroup            string `json:"experimentGroup,omitempty"`
	UserLeftExperiment         bool   `json:"userLeftExperiment,omitempty"`
	IsGuest                    bool   `json:"isGuest,omitempty"`
	IsInternal                 bool   `json:"isInternal,omitempty"`
	IsJBTeam                   bool   `json:"isJBTeam,omitempty"`
	FolderSelected             bool   `json:"folderSelected,omitempty"`
	ProjectFolderSelected      bool   `json:"projectFolderSelected,omitempty"`
	SavedQueryFolderSelected   bool   `json:"savedQueryFolderSelected,omitempty"`
	TagFolderSelected          bool   `json:"tagFolderSelected,omitempty"`
	IsSingleIssue              bool   `json:"isSingleIssue,omitempty"`
	IsTreeView                 bool   `json:"isTreeView,omitempty"`
	QueryComputationTimeMs     int64  `json:"queryComputationTimeMs,omitempty"`
	TimeSinceLastSearchMs      int64  `json:"timeSinceLastSearchMs,omitempty"`
	RecentSearchesCount        int    `json:"recentSearchesCount,omitempty"`
	LastSearchQueryTextSimilarity float64 `json:"lastSearchQueryTextSimilarity,omitempty"`
	LastSearchQueryTokenSimilarity float64 `json:"lastSearchQueryTokenSimilarity,omitempty"`
}

// SortedIssue تذكرة مرتبة في نتائج البحث
type SortedIssue struct {
	ID             string              `json:"id"`
	Summary        string              `json:"summary,omitempty"`
	Description    string              `json:"description,omitempty"`
	Resolved       *int64              `json:"resolved,omitempty"`
	Created        int64               `json:"created,omitempty"`
	Updated        int64               `json:"updated,omitempty"`
	Reporter       *User               `json:"reporter,omitempty"`
	Updater        *User               `json:"updater,omitempty"`
	Project        *Project            `json:"project,omitempty"`
	Fields         []*IssueCustomField `json:"fields,omitempty"`
	Tags           []*Tag              `json:"tags,omitempty"`
	Visibility     *Visibility         `json:"visibility,omitempty"`
	Votes          int                 `json:"votes,omitempty"`
	Watchers       *IssueWatchers      `json:"watchers,omitempty"`
	UsersTyping    []*UserTyping       `json:"usersTyping,omitempty"`
	CanUndoComment bool                `json:"canUndoComment,omitempty"`
	CanAddPublicComment bool           `json:"canAddPublicComment,omitempty"`
	Type           string              `json:"$type,omitempty"`
}
