package app

import (
	"errors"

	"github.com/jackc/pgx/v5"

	"youtrack_backend/channels/model"
	"youtrack_backend/channels/model/fields"
	"youtrack_backend/channels/request"
)

// SubscribeIssueList ينشئ اشتراكًا في قائمة المشاكل مع الجلب الانتقائي.
func (a *YouTrackApp) SubscribeIssueList(c request.CTX, userID string, req *model.IssueListSubscriptionRequest, tree *fields.FieldTree) (*model.IssueListSubscriptionBean, *model.AppError) {
	if userID == "" {
		userID = c.UserID()
	}
	ctx := c.Context()
	sub, err := a.Store().Subscriptions().SubscribeIssueList(ctx, userID, req, tree)
	if err != nil {
		return nil, model.NewInternalError("App.SubscribeIssueList", "failed to subscribe to issue list", err)
	}
	return sub, nil
}

// GetIssueListSubscriptionByTicket يجلب اشتراكًا حسب تذكرته مع الجلب الانتقائي.
func (a *YouTrackApp) GetIssueListSubscriptionByTicket(c request.CTX, userID string, ticket string, tree *fields.FieldTree) (*model.IssueListSubscriptionBean, *model.AppError) {
	if userID == "" {
		userID = c.UserID()
	}
	ctx := c.Context()
	sub, err := a.Store().Subscriptions().GetIssueListSubscriptionByTicket(ctx, userID, ticket, tree)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.NewNotFoundError("App.GetIssueListSubscriptionByTicket", "issue list subscription not found")
		}
		return nil, model.NewInternalError("App.GetIssueListSubscriptionByTicket", "failed to load issue list subscription", err)
	}
	return sub, nil
}
