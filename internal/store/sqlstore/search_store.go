package sqlstore

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"youtrack_backend/internal/api/fields"
	"youtrack_backend/internal/model"
)

// SearchStore هو تنفيذ SearchStore باستخدام PostgreSQL.
type SearchStore struct {
	db *pgxpool.Pool
}

// GetAssist يجلب بيانات مساعد البحث مع دعم الجلب الانتقائي للحقول.
func (s *SearchStore) GetAssist(ctx context.Context, query string, caret int, tree *fields.FieldTree) (*model.SearchAssistResponse, error) {
	resp := &model.SearchAssistResponse{
		Query: query,
		Caret: caret,
		Type:  "SearchAssistResponse",
	}

	// 1. جلب الاستجابة الأساسية
	var responseID int
	err := s.db.QueryRow(ctx, `
		SELECT id, ignore_unresolved_setting
		FROM search_assist_responses
		WHERE query = $1 AND caret = $2
		LIMIT 1`, query, caret).Scan(&responseID, &resp.IgnoreUnresolvedSetting)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			// إذا لم يوجد استجابة مخزنة، نعيد استجابة فارغة افتراضية
			return resp, nil
		}
		return nil, err
	}
	resp.ID = strconv.Itoa(responseID)

	// Note: IDs in database are SERIAL (int), but model.ID is string.
	// I'll adjust the scan or the model if needed, but for now I'll handle it carefully.
	// Actually search_assist_responses.id is SERIAL.

	// 2. StyleRanges
	if tree == nil || tree.Has("styleRanges") {
		rows, err := s.db.Query(ctx, `
			SELECT length, start_pos, style, COALESCE(title, '')
			FROM search_style_ranges
			WHERE response_id = $1`, responseID)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				sr := &model.SearchStyleRange{Type: "SearchStyleRange"}
				if err := rows.Scan(&sr.Length, &sr.Start, &sr.Style, &sr.Title); err == nil {
					resp.StyleRanges = append(resp.StyleRanges, sr)
				}
			}
		}
	}

	// 3. Suggestions
	if tree == nil || tree.Has("suggestions") {
		rows, err := s.db.Query(ctx, `
			SELECT description, suggestion_group, icon, suggestion_option, prefix, suffix,
			       class_name, matching_start, matching_end, caret, completion_start, completion_end
			FROM search_suggestions
			WHERE response_id = $1`, responseID)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				sg := &model.SearchSuggestion{Type: "SearchSuggestion"}
				if err := rows.Scan(&sg.Description, &sg.Group, &sg.Icon, &sg.Option, &sg.Prefix, &sg.Suffix,
					&sg.ClassName, &sg.MatchingStart, &sg.MatchingEnd, &sg.Caret, &sg.CompletionStart, &sg.CompletionEnd); err == nil {
					resp.Suggestions = append(resp.Suggestions, sg)
				}
			}
		}
	}

	// 4. SortProperties
	if tree == nil || tree.Has("sortProperties") {
		rows, err := s.db.Query(ctx, `
			SELECT p.property_id, p.asc_order, f.id, f.name, COALESCE(f.sortable_presentation, '')
			FROM search_sort_properties p
			LEFT JOIN search_sort_fields f ON p.sort_field_id = f.id
			WHERE p.response_id = $1`, responseID)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				sp := &model.SearchSortProperty{Type: "SortProperty"}
				sf := &model.SearchSortField{Type: "SortField"}
				if err := rows.Scan(&sp.ID, &sp.Asc, &sf.ID, &sf.Name, &sf.SortablePresentation); err == nil {
					sp.SortField = sf
					resp.SortProperties = append(resp.SortProperties, sp)
				}
			}
		}
	}

	// 5. QueryFeatures
	if tree == nil || tree.Has("queryFeatures") {
		qf := &model.SearchQueryFeatures{Type: "SearchQueryFeatures"}
		err := s.db.QueryRow(ctx, `
			SELECT contains_wildcard, contains_latin, contains_digits, contains_not_latin, number_of_words,
			       has_and, has_or, has_sorting, query_length, fields_count, user_in_fields, me_in_fields,
			       date_period_in_fields, sorted_by_relevance, sort_by_relevance_setting, experiment_version,
			       experiment_group, user_left_experiment, is_guest, is_internal, is_jb_team, folder_selected,
			       project_folder_selected, saved_query_folder_selected, tag_folder_selected, is_single_issue,
			       is_tree_view, query_computation_time_ms, time_since_last_search_ms, recent_searches_count,
			       last_search_query_text_similarity, last_search_query_token_similarity,
			       predefined_field_comment_text_field, predefined_field_work_text_field, predefined_field_vcs_changes_field,
			       predefined_field_ticket_cc_groups_field, predefined_field_mentions_field, predefined_field_underestimation_field,
			       predefined_field_saved_query_field, predefined_field_content_field, predefined_field_project_field,
			       predefined_field_star_field, predefined_field_attachment_name_field, predefined_field_voted_by_field,
			       predefined_field_article_field, predefined_field_reaction_from_field, predefined_field_by_field,
			       predefined_field_commented_by_field, predefined_field_custom_field, predefined_field_code_field,
			       predefined_field_ticket_cc_field, predefined_field_summary_field, predefined_field_mentioned_in_field,
			       predefined_field_organization_field, predefined_field_issue_field, predefined_field_votes_field,
			       predefined_field_attachments_field, predefined_field_links_field, predefined_field_title_field,
			       predefined_field_work_field, predefined_field_attachment_text_field, predefined_field_comments_field,
			       predefined_field_document_type_field, predefined_field_action_field, predefined_field_has_field,
			       predefined_field_similar_to_field, predefined_field_tag_field, predefined_field_description_field,
			       predefined_field_visible_to_field, predefined_field_sort_by_field, predefined_field_commented_field,
			       predefined_field_article_author_field, predefined_field_updated_by_field, predefined_field_submitted_by_field,
			       predefined_field_resolved_field, predefined_field_created_field, predefined_field_updated_field
			FROM search_query_features
			WHERE response_id = $1`, responseID).Scan(
			&qf.ContainsWildcard, &qf.ContainsLatin, &qf.ContainsDigits, &qf.ContainsNotLatin, &qf.NumberOfWords,
			&qf.HasAnd, &qf.HasOr, &qf.HasSorting, &qf.QueryLength, &qf.FieldsCount, &qf.UserInFields, &qf.MeInFields,
			&qf.DatePeriodInFields, &qf.SortedByRelevance, &qf.SortByRelevanceSetting, &qf.ExperimentVersion,
			&qf.ExperimentGroup, &qf.UserLeftExperiment, &qf.IsGuest, &qf.IsInternal, &qf.IsJBTeam, &qf.FolderSelected,
			&qf.ProjectFolderSelected, &qf.SavedQueryFolderSelected, &qf.TagFolderSelected, &qf.IsSingleIssue,
			&qf.IsTreeView, &qf.QueryComputationTimeMs, &qf.TimeSinceLastSearchMs, &qf.RecentSearchesCount,
			&qf.LastSearchQueryTextSimilarity, &qf.LastSearchQueryTokenSimilarity,
			&qf.PredefinedFieldCommentTextField, &qf.PredefinedFieldWorkTextField, &qf.PredefinedFieldVcsChangesField,
			&qf.PredefinedFieldTicketCCGroupsField, &qf.PredefinedFieldMentionsField, &qf.PredefinedFieldUnderestimationField,
			&qf.PredefinedFieldSavedQueryField, &qf.PredefinedFieldContentField, &qf.PredefinedFieldProjectField,
			&qf.PredefinedFieldStarField, &qf.PredefinedFieldAttachmentNameField, &qf.PredefinedFieldVotedByField,
			&qf.PredefinedFieldArticleField, &qf.PredefinedFieldReactionFromField, &qf.PredefinedFieldByField,
			&qf.PredefinedFieldCommentedByField, &qf.PredefinedFieldCustomField, &qf.PredefinedFieldCodeField,
			&qf.PredefinedFieldTicketCCField, &qf.PredefinedFieldSummaryField, &qf.PredefinedFieldMentionedInField,
			&qf.PredefinedFieldOrganizationField, &qf.PredefinedFieldIssueField, &qf.PredefinedFieldVotesField,
			&qf.PredefinedFieldAttachmentsField, &qf.PredefinedFieldLinksField, &qf.PredefinedFieldTitleField,
			&qf.PredefinedFieldWorkField, &qf.PredefinedFieldAttachmentTextField, &qf.PredefinedFieldCommentsField,
			&qf.PredefinedFieldDocumentTypeField, &qf.PredefinedFieldActionField, &qf.PredefinedFieldHasField,
			&qf.PredefinedFieldSimilarToField, &qf.PredefinedFieldTagField, &qf.PredefinedFieldDescriptionField,
			&qf.PredefinedFieldVisibleToField, &qf.PredefinedFieldSortByField, &qf.PredefinedFieldCommentedField,
			&qf.PredefinedFieldArticleAuthorField, &qf.PredefinedFieldUpdatedByField, &qf.PredefinedFieldSubmittedByField,
			&qf.PredefinedFieldResolvedField, &qf.PredefinedFieldCreatedField, &qf.PredefinedFieldUpdatedField,
		)
		if err == nil {
			resp.QueryFeatures = qf
		}
	}

	return resp, nil
}
