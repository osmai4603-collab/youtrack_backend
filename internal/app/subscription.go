package app

import (
	"context"

	"github.com/jackc/pgx/v5"

	"youtrack_backend/internal/api/fields"
	"youtrack_backend/internal/model"
)

// SubscribeIssueList ينشئ اشتراكًا في قائمة المشاكل مع الجلب الانتقائي حسب
// شجرة الحقول (مطابق لـ request18.txt).
func (a *App) SubscribeIssueList(ctx context.Context, userID string, req *model.IssueListSubscriptionRequest, tree *fields.FieldTree) (*model.IssueListSubscriptionBean, error) {
	sub, err := a.store.Subscriptions().SubscribeIssueList(ctx, userID, req, tree)
	if err != nil {
		return nil, model.Internal("failed to subscribe to issue list: %v", err)
	}
	return sub, nil
}

// GetIssueListSubscriptionByTicket يجلب اشتراكًا حسب تذكرته مع الجلب الانتقائي.
func (a *App) GetIssueListSubscriptionByTicket(ctx context.Context, userID string, ticket string, tree *fields.FieldTree) (*model.IssueListSubscriptionBean, error) {
	sub, err := a.store.Subscriptions().GetIssueListSubscriptionByTicket(ctx, userID, ticket, tree)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, model.NotFound("issue list subscription not found")
		}
		return nil, model.Internal("failed to load issue list subscription: %v", err)
	}
	return sub, nil
}
