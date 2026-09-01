package model

// SearchAssistResponse يمثّل الاستجابة لنقطة النهاية api/search/assist.
type SearchAssistResponse struct {
	ID                       string               `json:"id,omitempty"`
	Caret                    int                  `json:"caret"`
	Query                    string               `json:"query"`
	StyleRanges              []*SearchStyleRange  `json:"styleRanges,omitempty"`
	Suggestions              []*SearchSuggestion  `json:"suggestions,omitempty"`
	SortProperties           []*SearchSortProperty `json:"sortProperties,omitempty"`
	IgnoreUnresolvedSetting  bool                 `json:"ignoreUnresolvedSetting"`
	QueryFeatures            *SearchQueryFeatures `json:"queryFeatures,omitempty"`
	AST                      *SearchAST           `json:"ast,omitempty"`
	Type                     string               `json:"$type,omitempty"`
}

// SearchStyleRange يصف نمط جزء من استعلام البحث.
type SearchStyleRange struct {
	Length int    `json:"length"`
	Start  int    `json:"start"`
	Style  string `json:"style"`
	Title  string `json:"title,omitempty"`
	Type   string `json:"$type,omitempty"`
}

// SearchSuggestion يمثّل اقتراحاً للإكمال التلقائي في البحث.
type SearchSuggestion struct {
	Description     string `json:"description,omitempty"`
	Group           string `json:"group,omitempty"`
	Icon            string `json:"icon,omitempty"`
	Option          string `json:"option,omitempty"`
	Prefix          string `json:"prefix,omitempty"`
	Suffix          string `json:"suffix,omitempty"`
	ClassName       string `json:"className,omitempty"`
	MatchingStart   int    `json:"matchingStart"`
	MatchingEnd     int    `json:"matchingEnd"`
	Caret           int    `json:"caret"`
	CompletionStart int    `json:"completionStart"`
	CompletionEnd   int    `json:"completionEnd"`
	Type            string `json:"$type,omitempty"`
}

// SearchSortProperty يصف خاصية الترتيب في البحث.
type SearchSortProperty struct {
	ID        string           `json:"id"`
	Asc       bool             `json:"asc"`
	SortField *SearchSortField `json:"sortField,omitempty"`
	Type      string           `json:"$type,omitempty"`
}

// SearchSortField يصف حقل الترتيب.
type SearchSortField struct {
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	SortablePresentation string `json:"sortablePresentation,omitempty"`
	Type                 string `json:"$type,omitempty"`
}

// SearchQueryFeatures يحتوي على خصائص وإحصائيات حول استعلام البحث.
type SearchQueryFeatures struct {
	ContainsWildcard                 bool    `json:"containsWildcard"`
	ContainsLatin                    bool    `json:"containsLatin"`
	ContainsDigits                   bool    `json:"containsDigits"`
	ContainsNotLatin                 bool    `json:"containsNotLatin"`
	NumberOfWords                    int     `json:"numberOfWords"`
	HasAnd                           bool    `json:"hasAnd"`
	HasOr                            bool    `json:"hasOr"`
	HasSorting                       bool    `json:"hasSorting"`
	Length                           int     `json:"length"`
	QueryLength                      int     `json:"queryLength"`
	FieldsCount                      int     `json:"fieldsCount"`
	UserInFields                     bool    `json:"userInFields"`
	MeInFields                       bool    `json:"meInFields"`
	DatePeriodInFields               bool    `json:"datePeriodInFields"`
	SortedByRelevance                bool    `json:"sortedByRelevance"`
	SortByRelevanceSetting           bool    `json:"sortByRelevanceSetting"`
	ExperimentVersion                string  `json:"experimentVersion"`
	ExperimentGroup                  string  `json:"experimentGroup"`
	UserLeftExperiment               bool    `json:"userLeftExperiment"`
	IsGuest                          bool    `json:"isGuest"`
	IsInternal                       bool    `json:"isInternal"`
	IsJBTeam                         bool    `json:"isJBTeam"`
	FolderSelected                   bool    `json:"folderSelected"`
	ProjectFolderSelected            bool    `json:"projectFolderSelected"`
	SavedQueryFolderSelected         bool    `json:"savedQueryFolderSelected"`
	TagFolderSelected                bool    `json:"tagFolderSelected"`
	IsSingleIssue                    bool    `json:"isSingleIssue"`
	IsTreeView                       bool    `json:"isTreeView"`
	QueryComputationTimeMs           int64   `json:"queryComputationTimeMs"`
	TimeSinceLastSearchMs            int64   `json:"timeSinceLastSearchMs"`
	RecentSearchesCount              int     `json:"recentSearchesCount"`
	LastSearchQueryTextSimilarity    float64 `json:"lastSearchQueryTextSimilarity"`
	LastSearchQueryTokenSimilarity   float64 `json:"lastSearchQueryTokenSimilarity"`
	PredefinedFieldCommentTextField  bool    `json:"PredefinedFieldCommentTextField"`
	PredefinedFieldWorkTextField     bool    `json:"PredefinedFieldWorkTextField"`
	PredefinedFieldVcsChangesField   bool    `json:"PredefinedFieldVcsChangesField"`
	PredefinedFieldTicketCCGroupsField bool  `json:"PredefinedFieldTicketCCGroupsField"`
	PredefinedFieldMentionsField     bool    `json:"PredefinedFieldMentionsField"`
	PredefinedFieldUnderestimationField bool `json:"PredefinedFieldUnderestimationField"`
	PredefinedFieldSavedQueryField   bool    `json:"PredefinedFieldSavedQueryField"`
	PredefinedFieldContentField      bool    `json:"PredefinedFieldContentField"`
	PredefinedFieldProjectField      bool    `json:"PredefinedFieldProjectField"`
	PredefinedFieldStarField         bool    `json:"PredefinedFieldStarField"`
	PredefinedFieldAttachmentNameField bool  `json:"PredefinedFieldAttachmentNameField"`
	PredefinedFieldVotedByField      bool    `json:"PredefinedFieldVotedByField"`
	PredefinedFieldArticleField      bool    `json:"PredefinedFieldArticleField"`
	PredefinedFieldReactionFromField bool    `json:"PredefinedFieldReactionFromField"`
	PredefinedFieldByField           bool    `json:"PredefinedFieldByField"`
	PredefinedFieldCommentedByField  bool    `json:"PredefinedFieldCommentedByField"`
	PredefinedFieldCustomField       bool    `json:"PredefinedFieldCustomField"`
	PredefinedFieldCodeField         bool    `json:"PredefinedFieldCodeField"`
	PredefinedFieldTicketCCField     bool    `json:"PredefinedFieldTicketCCField"`
	PredefinedFieldSummaryField      bool    `json:"PredefinedFieldSummaryField"`
	PredefinedFieldMentionedInField  bool    `json:"PredefinedFieldMentionedInField"`
	PredefinedFieldOrganizationField bool    `json:"PredefinedFieldOrganizationField"`
	PredefinedFieldIssueField        bool    `json:"PredefinedFieldIssueField"`
	PredefinedFieldVotesField        bool    `json:"PredefinedFieldVotesField"`
	PredefinedFieldAttachmentsField  bool    `json:"PredefinedFieldAttachmentsField"`
	PredefinedFieldLinksField        bool    `json:"PredefinedFieldLinksField"`
	PredefinedFieldTitleField        bool    `json:"PredefinedFieldTitleField"`
	PredefinedFieldWorkField         bool    `json:"PredefinedFieldWorkField"`
	PredefinedFieldAttachmentTextField bool  `json:"PredefinedFieldAttachmentTextField"`
	PredefinedFieldCommentsField     bool    `json:"PredefinedFieldCommentsField"`
	PredefinedFieldDocumentTypeField bool    `json:"PredefinedFieldDocumentTypeField"`
	PredefinedFieldActionField       bool    `json:"PredefinedFieldActionField"`
	PredefinedFieldHasField          bool    `json:"PredefinedFieldHasField"`
	PredefinedFieldSimilarToField    bool    `json:"PredefinedFieldSimilarToField"`
	PredefinedFieldTagField          bool    `json:"PredefinedFieldTagField"`
	PredefinedFieldDescriptionField  bool    `json:"PredefinedFieldDescriptionField"`
	PredefinedFieldVisibleToField    bool    `json:"PredefinedFieldVisibleToField"`
	PredefinedFieldSortByField       bool    `json:"PredefinedFieldSortByField"`
	PredefinedFieldCommentedField    bool    `json:"PredefinedFieldCommentedField"`
	PredefinedFieldArticleAuthorField bool   `json:"PredefinedFieldArticleAuthorField"`
	PredefinedFieldUpdatedByField    bool    `json:"PredefinedFieldUpdatedByField"`
	PredefinedFieldSubmittedByField  bool    `json:"PredefinedFieldSubmittedByField"`
	PredefinedFieldResolvedField     bool    `json:"PredefinedFieldResolvedField"`
	PredefinedFieldCreatedField      bool    `json:"PredefinedFieldCreatedField"`
	PredefinedFieldUpdatedField      bool    `json:"PredefinedFieldUpdatedField"`
	ID                               string  `json:"id,omitempty"`
	Type                             string  `json:"$type,omitempty"`
}

// SearchAST يمثّل شجرة التحليل النحوي لاستعلام البحث.
type SearchAST struct {
	Expression *SearchExpression `json:"expression,omitempty"`
	Type       string            `json:"$type,omitempty"`
}

// SearchExpression يمثّل تعبيراً في الـ AST.
type SearchExpression struct {
	Operator string          `json:"operator,omitempty"`
	Left     *SearchExpression `json:"left,omitempty"`
	Right    *SearchExpression `json:"right,omitempty"`
	Terms    []*SearchTerm   `json:"terms,omitempty"`
	Type     string          `json:"$type,omitempty"`
}

// SearchTerm يمثّل مصطلحاً في التعبير.
type SearchTerm struct {
	Text   string         `json:"text,omitempty"`
	Minus  bool           `json:"minus,omitempty"`
	Value  *SearchValue   `json:"value,omitempty"`
	Field  *SearchField   `json:"field,omitempty"`
	Values []*SearchValue `json:"values,omitempty"`
}

// SearchValue يمثّل قيمة في مصطلح البحث.
type SearchValue struct {
	Name       string            `json:"name,omitempty"`
	Minus      bool              `json:"minus,omitempty"`
	Start      int               `json:"start,omitempty"`
	Entity     *SearchEntity     `json:"entity,omitempty"`
	Field      *SearchField      `json:"field,omitempty"`
	Left       *SearchValue      `json:"left,omitempty"`
	Right      *SearchValue      `json:"right,omitempty"`
	Expression *SearchExpression `json:"expression,omitempty"`
	Type       string            `json:"$type,omitempty"`
}

// SearchField يمثّل حقلاً في مصطلح البحث.
type SearchField struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// SearchEntity يمثّل كياناً (مثل مستخدم) في قيمة البحث.
type SearchEntity struct {
	ID        string `json:"id,omitempty"`
	Login     string `json:"login,omitempty"`
	Email     string `json:"email,omitempty"`
	AvatarURL string `json:"avatarUrl,omitempty"`
	Name      string `json:"name,omitempty"`
}
