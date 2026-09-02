package timerlayer

import (
	"context"
	"time"

	"youtrack_backend/channels/model"
	"youtrack_backend/channels/model/fields"
	"youtrack_backend/channels/store"
	"youtrack_backend/channels/model/shared/mlog"
)

// TimerLayer يغلّف Store ويقيس زمن كل استعلام.
type TimerLayer struct {
	store.Store
}

func New(childStore store.Store) *TimerLayer {
	return &TimerLayer{Store: childStore}
}

func (s *TimerLayer) Users() store.UserStore {
	return &TimerLayerUserStore{UserStore: s.Store.Users()}
}

// TimerLayerUserStore يغلّف UserStore.
type TimerLayerUserStore struct {
	store.UserStore
}

func (s *TimerLayerUserStore) GetByID(ctx context.Context, id string) (*model.User, error) {
	start := time.Now()
	result, err := s.UserStore.GetByID(ctx, id)
	elapsed := time.Since(start)
	if elapsed > 200*time.Millisecond {
		mlog.Warn("[SLOW QUERY] UserStore.GetByID took " + elapsed.String())
	}
	return result, err
}

func (s *TimerLayerUserStore) GetByLogin(ctx context.Context, login string) (*model.User, error) {
	start := time.Now()
	result, err := s.UserStore.GetByLogin(ctx, login)
	elapsed := time.Since(start)
	if elapsed > 200*time.Millisecond {
		mlog.Warn("[SLOW QUERY] UserStore.GetByLogin took " + elapsed.String())
	}
	return result, err
}

func (s *TimerLayerUserStore) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	start := time.Now()
	result, err := s.UserStore.GetByEmail(ctx, email)
	elapsed := time.Since(start)
	if elapsed > 200*time.Millisecond {
		mlog.Warn("[SLOW QUERY] UserStore.GetByEmail took " + elapsed.String())
	}
	return result, err
}

func (s *TimerLayerUserStore) Create(ctx context.Context, u *model.User) (error) {
	start := time.Now()
	err := s.UserStore.Create(ctx, u)
	elapsed := time.Since(start)
	if elapsed > 200*time.Millisecond {
		mlog.Warn("[SLOW QUERY] UserStore.Create took " + elapsed.String())
	}
	return err
}

func (s *TimerLayerUserStore) All(ctx context.Context) ([]*model.User, error) {
	start := time.Now()
	result, err := s.UserStore.All(ctx)
	elapsed := time.Since(start)
	if elapsed > 200*time.Millisecond {
		mlog.Warn("[SLOW QUERY] UserStore.All took " + elapsed.String())
	}
	return result, err
}

func (s *TimerLayerUserStore) GetProfile(ctx context.Context, userID string) (*model.UserProfile, error) {
	start := time.Now()
	result, err := s.UserStore.GetProfile(ctx, userID)
	elapsed := time.Since(start)
	if elapsed > 200*time.Millisecond {
		mlog.Warn("[SLOW QUERY] UserStore.GetProfile took " + elapsed.String())
	}
	return result, err
}

func (s *TimerLayerUserStore) CreateProfile(ctx context.Context, p *model.UserProfile) (error) {
	start := time.Now()
	err := s.UserStore.CreateProfile(ctx, p)
	elapsed := time.Since(start)
	if elapsed > 200*time.Millisecond {
		mlog.Warn("[SLOW QUERY] UserStore.CreateProfile took " + elapsed.String())
	}
	return err
}

func (s *TimerLayerUserStore) GetMe(ctx context.Context, userID string, tree *fields.FieldTree) (*model.User, error) {
	start := time.Now()
	result, err := s.UserStore.GetMe(ctx, userID, tree)
	elapsed := time.Since(start)
	if elapsed > 200*time.Millisecond {
		mlog.Warn("[SLOW QUERY] UserStore.GetMe took " + elapsed.String())
	}
	return result, err
}

func (s *TimerLayerUserStore) GetFeatureFlags(ctx context.Context) ([]*model.FeatureFlag, error) {
	start := time.Now()
	result, err := s.UserStore.GetFeatureFlags(ctx)
	elapsed := time.Since(start)
	if elapsed > 200*time.Millisecond {
		mlog.Warn("[SLOW QUERY] UserStore.GetFeatureFlags took " + elapsed.String())
	}
	return result, err
}

func (s *TimerLayerUserStore) GetRecentIssues(ctx context.Context, userID string, limit int, offset int) ([]*model.RecentIssue, error) {
	start := time.Now()
	result, err := s.UserStore.GetRecentIssues(ctx, userID, limit, offset)
	elapsed := time.Since(start)
	if elapsed > 200*time.Millisecond {
		mlog.Warn("[SLOW QUERY] UserStore.GetRecentIssues took " + elapsed.String())
	}
	return result, err
}

func (s *TimerLayerUserStore) GetRecentArticles(ctx context.Context, userID string, limit int, offset int) ([]*model.RecentArticle, error) {
	start := time.Now()
	result, err := s.UserStore.GetRecentArticles(ctx, userID, limit, offset)
	elapsed := time.Since(start)
	if elapsed > 200*time.Millisecond {
		mlog.Warn("[SLOW QUERY] UserStore.GetRecentArticles took " + elapsed.String())
	}
	return result, err
}

func (s *TimerLayerUserStore) GetGrazieProfile(ctx context.Context, userID string) (*model.GrazieUserProfile, error) {
	start := time.Now()
	result, err := s.UserStore.GetGrazieProfile(ctx, userID)
	elapsed := time.Since(start)
	if elapsed > 200*time.Millisecond {
		mlog.Warn("[SLOW QUERY] UserStore.GetGrazieProfile took " + elapsed.String())
	}
	return result, err
}

func (s *TimerLayerUserStore) GetGeneralProfile(ctx context.Context, userID string) (*model.GeneralUserProfile, error) {
	start := time.Now()
	result, err := s.UserStore.GetGeneralProfile(ctx, userID)
	elapsed := time.Since(start)
	if elapsed > 200*time.Millisecond {
		mlog.Warn("[SLOW QUERY] UserStore.GetGeneralProfile took " + elapsed.String())
	}
	return result, err
}

func (s *TimerLayerUserStore) GetQuestionnaireProfile(ctx context.Context, userID string) (*model.QuestionnaireUserProfile, error) {
	start := time.Now()
	result, err := s.UserStore.GetQuestionnaireProfile(ctx, userID)
	elapsed := time.Since(start)
	if elapsed > 200*time.Millisecond {
		mlog.Warn("[SLOW QUERY] UserStore.GetQuestionnaireProfile took " + elapsed.String())
	}
	return result, err
}

func (s *TimerLayerUserStore) GetHubMe(ctx context.Context, userID string) (*model.HubUser, error) {
	start := time.Now()
	result, err := s.UserStore.GetHubMe(ctx, userID)
	elapsed := time.Since(start)
	if elapsed > 200*time.Millisecond {
		mlog.Warn("[SLOW QUERY] UserStore.GetHubMe took " + elapsed.String())
	}
	return result, err
}

func (s *TimerLayerUserStore) InboxFolders(ctx context.Context, userID string) ([]*model.InboxFolder, error) {
	start := time.Now()
	result, err := s.UserStore.InboxFolders(ctx, userID)
	elapsed := time.Since(start)
	if elapsed > 200*time.Millisecond {
		mlog.Warn("[SLOW QUERY] UserStore.InboxFolders took " + elapsed.String())
	}
	return result, err
}
