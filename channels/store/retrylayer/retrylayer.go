package retrylayer

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"youtrack_backend/channels/model"
	"youtrack_backend/channels/model/fields"
	"youtrack_backend/channels/store"
)

const maxRetries = 3

// RetryLayer يغلّف Store ويعيد محاولة الاستعلامات عند الفشل المؤقت.
type RetryLayer struct {
	store.Store
}

func New(childStore store.Store) *RetryLayer {
	return &RetryLayer{Store: childStore}
}

func (s *RetryLayer) Users() store.UserStore {
	return &RetryLayerUserStore{UserStore: s.Store.Users()}
}

// isRetryable يحدد ما إذا كان الخطأ قابلاً لإعادة المحاولة.
func isRetryable(err error) bool {
	if err == nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "40P01" || pgErr.Code == "55P03" || len(pgErr.Code) >= 2 && pgErr.Code[:2] == "08"
	}
	var netErr net.Error
	return errors.As(err, &netErr) && (netErr.Timeout() || netErr.Temporary())
}

// retry يعيد المحاولة مع backoff أسّي.
func retry[T any](ctx context.Context, fn func() (T, error)) (T, error) {
	var result T
	var err error
	for i := 0; i < maxRetries; i++ {
		result, err = fn()
		if err == nil {
			return result, nil
		}
		if !isRetryable(err) {
			return result, err
		}
		if err := waitBeforeRetry(ctx, i); err != nil {
			return result, err
		}
	}
	return result, err
}

func retryNoRet(ctx context.Context, fn func() error) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		if !isRetryable(err) {
			return err
		}
		if waitErr := waitBeforeRetry(ctx, i); waitErr != nil {
			return waitErr
		}
	}
	return err
}

func waitBeforeRetry(ctx context.Context, attempt int) error {
	delay := time.Duration(attempt+1) * 100 * time.Millisecond
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// RetryLayerUserStore يغلّف UserStore.
type RetryLayerUserStore struct {
	store.UserStore
}

func (s *RetryLayerUserStore) GetByID(ctx context.Context, id string) (*model.User, error) {
	return retry(ctx, func() (*model.User, error) {
		return s.UserStore.GetByID(ctx, id)
	})
}

func (s *RetryLayerUserStore) GetByLogin(ctx context.Context, login string) (*model.User, error) {
	return retry(ctx, func() (*model.User, error) {
		return s.UserStore.GetByLogin(ctx, login)
	})
}

func (s *RetryLayerUserStore) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	return retry(ctx, func() (*model.User, error) {
		return s.UserStore.GetByEmail(ctx, email)
	})
}

func (s *RetryLayerUserStore) Create(ctx context.Context, u *model.User) error {
	return retryNoRet(ctx, func() error {
		return s.UserStore.Create(ctx, u)
	})
}

func (s *RetryLayerUserStore) All(ctx context.Context) ([]*model.User, error) {
	return retry(ctx, func() ([]*model.User, error) {
		return s.UserStore.All(ctx)
	})
}

func (s *RetryLayerUserStore) GetProfile(ctx context.Context, userID string) (*model.UserProfile, error) {
	return retry(ctx, func() (*model.UserProfile, error) {
		return s.UserStore.GetProfile(ctx, userID)
	})
}

func (s *RetryLayerUserStore) CreateProfile(ctx context.Context, p *model.UserProfile) error {
	return retryNoRet(ctx, func() error {
		return s.UserStore.CreateProfile(ctx, p)
	})
}

func (s *RetryLayerUserStore) GetMe(ctx context.Context, userID string, tree *fields.FieldTree) (*model.User, error) {
	return retry(ctx, func() (*model.User, error) {
		return s.UserStore.GetMe(ctx, userID, tree)
	})
}

func (s *RetryLayerUserStore) GetFeatureFlags(ctx context.Context) ([]*model.FeatureFlag, error) {
	return retry(ctx, func() ([]*model.FeatureFlag, error) {
		return s.UserStore.GetFeatureFlags(ctx)
	})
}

func (s *RetryLayerUserStore) GetRecentIssues(ctx context.Context, userID string, limit int, offset int) ([]*model.RecentIssue, error) {
	return retry(ctx, func() ([]*model.RecentIssue, error) {
		return s.UserStore.GetRecentIssues(ctx, userID, limit, offset)
	})
}

func (s *RetryLayerUserStore) GetRecentArticles(ctx context.Context, userID string, limit int, offset int) ([]*model.RecentArticle, error) {
	return retry(ctx, func() ([]*model.RecentArticle, error) {
		return s.UserStore.GetRecentArticles(ctx, userID, limit, offset)
	})
}

func (s *RetryLayerUserStore) GetGrazieProfile(ctx context.Context, userID string) (*model.GrazieUserProfile, error) {
	return retry(ctx, func() (*model.GrazieUserProfile, error) {
		return s.UserStore.GetGrazieProfile(ctx, userID)
	})
}

func (s *RetryLayerUserStore) GetGeneralProfile(ctx context.Context, userID string) (*model.GeneralUserProfile, error) {
	return retry(ctx, func() (*model.GeneralUserProfile, error) {
		return s.UserStore.GetGeneralProfile(ctx, userID)
	})
}

func (s *RetryLayerUserStore) GetQuestionnaireProfile(ctx context.Context, userID string) (*model.QuestionnaireUserProfile, error) {
	return retry(ctx, func() (*model.QuestionnaireUserProfile, error) {
		return s.UserStore.GetQuestionnaireProfile(ctx, userID)
	})
}

func (s *RetryLayerUserStore) GetHubMe(ctx context.Context, userID string) (*model.HubUser, error) {
	return retry(ctx, func() (*model.HubUser, error) {
		return s.UserStore.GetHubMe(ctx, userID)
	})
}

func (s *RetryLayerUserStore) InboxFolders(ctx context.Context, userID string) ([]*model.InboxFolder, error) {
	return retry(ctx, func() ([]*model.InboxFolder, error) {
		return s.UserStore.InboxFolders(ctx, userID)
	})
}
