package model

// SortedIssuesResponse represents the response for /api/sortedIssues
type SortedIssuesResponse struct {
	Tree []*IssueTreeItem `json:"tree"`
	Type string           `json:"$type,omitempty"`
}

// IssueTreeItem represents an item in the sorted issues tree
type IssueTreeItem struct {
	ID                        string                  `json:"id"`
	SummaryTextSearchResult   *TextSearchResult        `json:"summaryTextSearchResult,omitempty"`
	DescriptionTextSearchResult *TextSearchResult      `json:"descriptionTextSearchResult,omitempty"`
	Matches                   bool                    `json:"matches"`
	Ordered                   bool                    `json:"ordered"`
	ParentID                  string                  `json:"parentId,omitempty"`
	SearchFeatures            *IssueSearchFeatures    `json:"searchFeatures,omitempty"`
	MLScore                   float64                 `json:"mlScore,omitempty"`
	Promoted                  bool                    `json:"promoted,omitempty"`
	Type                      string                  `json:"$type,omitempty"`
}

// TextSearchResult represents highlights in search results
type TextSearchResult struct {
	HighlightRanges []*HighlightRange `json:"highlightRanges,omitempty"`
	TextRange       *HighlightRange   `json:"textRange,omitempty"`
	Type            string            `json:"$type,omitempty"`
}

// HighlightRange represents a single highlight range
type HighlightRange struct {
	StartOffset int    `json:"startOffset"`
	EndOffset   int    `json:"endOffset"`
	Type        string `json:"$type,omitempty"`
}

// IssueSearchFeatures represents various metrics and features of an issue in search results
type IssueSearchFeatures struct {
	CommentsCount             int     `json:"commentsCount"`
	ExactTitleMatch           bool    `json:"exactTitleMatch"`
	FavoritesCount            int     `json:"favoritesCount"`
	TagsCount                 int     `json:"tagsCount"`
	ViewsCount                int     `json:"viewsCount"`
	VotesCount                int     `json:"votesCount"`
	QueryTitleWordIntersection int    `json:"queryTitleWordIntersection"`
	TimeSinceCommentedMs      int64   `json:"timeSinceCommentedMs"`
	IsStarred                 bool    `json:"isStarred"`
	CreatedDateMs             int64   `json:"createdDateMs"`
	TimeSinceCreatedMs        int64   `json:"timeSinceCreatedMs"`
	TimeSinceUpdatedMs        int64   `json:"timeSinceUpdatedMs"`
	TimeSinceResolvedMs       int64   `json:"timeSinceResolvedMs"`
	IsAuthor                  bool    `json:"isAuthor"`
	AuthorSameTeamsIntersection int   `json:"authorSameTeamsIntersection"`
	IsInFields                bool    `json:"isInFields"`
	IsAssignee                bool    `json:"isAssignee"`
	State                     string  `json:"state,omitempty"`
	Priority                  string  `json:"priority,omitempty"`
	IssueType                 string  `json:"type,omitempty"`
	IsInRelatedProject        bool    `json:"isInRelatedProject"`
	IsMentioned               bool    `json:"isMentioned"`
	IsUserUpdated             bool    `json:"isUserUpdated"`
	ContentLength             int     `json:"contentLength"`
	LastVisitedUsersCount     int     `json:"lastVisitedUsersCount"`
	LastVisited               int64   `json:"lastVisited"`
	LastVisitedOrder          int     `json:"lastVisitedOrder"`
	LastVisitedSameProject    bool    `json:"lastVisitedSameProject"`
	LastVisitedSameAuthor     bool    `json:"lastVisitedSameAuthor"`
	LastVisitedProjectIntersection int `json:"lastVisitedProjectIntersection"`
	LastVisitedAuthorIntersection  int `json:"lastVisitedAuthorIntersection"`
	TimeSinceLastVisitedMs    int64   `json:"timeSinceLastVisitedMs"`
	FullTextSearchScore       float64 `json:"fullTextSearchScore"`
	IsParent                  bool    `json:"isParent"`
	ChildrenCount             int     `json:"childrenCount"`
	IsChild                   bool    `json:"isChild"`
	SummaryHighlightsCount    int     `json:"summaryHighlightsCount"`
	SummaryHighlightsLength   int     `json:"summaryHighlightsLength"`
	SummaryHighlightsRatio    float64 `json:"summaryHighlightsRatio"`
	DescriptionHighlightsCount  int   `json:"descriptionHighlightsCount"`
	DescriptionHighlightsLength int   `json:"descriptionHighlightsLength"`
	DescriptionHighlightsRatio  float64 `json:"descriptionHighlightsRatio"`
	IsOriginalSearchMatch     bool    `json:"isOriginalSearchMatch"`
	IsLayoutSwitchMatch       bool    `json:"isLayoutSwitchMatch"`
	IsSuffixWildcardMatch     bool    `json:"isSuffixWildcardMatch"`
	IsFuzzyMatch              bool    `json:"isFuzzyMatch"`
	ComputationTimeMs         int64   `json:"computationTimeMs"`
	OriginalPosition          int     `json:"originalPosition"`
	IsInSprint                bool    `json:"isInSprint"`
	IsInActiveSprint          bool    `json:"isInActiveSprint"`
	PromotionKind             string  `json:"promotionKind,omitempty"`
	TimeSinceRecentNavigationMs int64 `json:"timeSinceRecentNavigationMs"`
	RecentlyNavigatedCount    int     `json:"recentlyNavigatedCount"`
	RecentlyNavigatedQueryTokenSimilarity float64 `json:"recentlyNavigatedQueryTokenSimilarity"`
	RecentlyNavigatedQueryTextSimilarity  float64 `json:"recentlyNavigatedQueryTextSimilarity"`
	TotalHits                 int     `json:"totalHits"`
	ID                        string  `json:"id"`
	Type                      string  `json:"$type,omitempty"`
}

func (r *SortedIssuesResponse) Normalize() {
	r.Type = "SortedIssuesResponse"
	for _, item := range r.Tree {
		item.Normalize()
	}
}

func (i *IssueTreeItem) Normalize() {
	i.Type = "IssueTreeItem"
	if i.SummaryTextSearchResult != nil {
		i.SummaryTextSearchResult.Type = "TextSearchResult"
		for _, r := range i.SummaryTextSearchResult.HighlightRanges {
			r.Type = "HighlightRange"
		}
	}
	if i.DescriptionTextSearchResult != nil {
		i.DescriptionTextSearchResult.Type = "TextSearchResult"
		for _, r := range i.DescriptionTextSearchResult.HighlightRanges {
			r.Type = "HighlightRange"
		}
	}
	if i.SearchFeatures != nil {
		i.SearchFeatures.Type = "IssueSearchFeatures"
	}
}
