package usecase

import (
	"youtrack_backend/internal/domain"
)

type UserUseCase struct {
	userRepo domain.UserRepository
}

func NewUserUseCase(userRepo domain.UserRepository) *UserUseCase {
	return &UserUseCase{
		userRepo: userRepo,
	}
}

func (uc *UserUseCase) GetCurrentUser(userID string) (*domain.User, error) {
	var user *domain.User

	if uc.userRepo != nil {
		dbUser, err := uc.userRepo.GetByID(userID)
		if err == nil && dbUser != nil {
			user = dbUser
		}
	}

	if user == nil {
		user = &domain.User{
			ID:              userID,
			Login:           "current_user",
			Email:           "user@example.com",
			Name:            "current_user",
			FullName:        "User Full Name",
			LocalizedName:   "User Full Name",
			AvatarURL:       "/hub/api/rest/avatar/" + userID,
			Online:          true,
			Banned:          false,
			CanReadProfile:  true,
			IsLocked:        false,
			IsEmailVerified: true,
			Guest:           false,
			UserType: &domain.UserType{
				ID:   "STANDARD",
				Name: "Standard",
				Type: "UserType",
			},
		}
	}

	if user.UserType == nil {
		user.UserType = &domain.UserType{
			ID:   "STANDARD",
			Name: "Standard",
			Type: "UserType",
		}
	}

	if user.Profiles == nil {
		user.Profiles = domain.DefaultUserProfiles()
	}

	if user.FeatureFlags == nil {
		user.FeatureFlags = []*domain.FeatureFlag{
			{ID: "jetbrains.youtrack.feature.generateAILinkSuggestions", Enabled: true, Type: "FeatureFlag"},
			{ID: "jetbrains.youtrack.feature.lazyActivityValues", Enabled: true, Type: "FeatureFlag"},
			{ID: "jetbrains.youtrack.feature.grazieSensitiveProjects", Enabled: true, Type: "FeatureFlag"},
			{ID: "jetbrains.youtrack.feature.teamcityLoadBuildParameters", Enabled: true, Type: "FeatureFlag"},
			{ID: "jetbrains.youtrack.feature.requireSpaceVCSBranchMonitoring", Enabled: true, Type: "FeatureFlag"},
			{ID: "jetbrains.youtrack.feature.spaceCommitReverseHashesRefactoringFlag", Enabled: true, Type: "FeatureFlag"},
			{ID: "jetbrains.youtrack.feature.commentReaction", Enabled: true, Type: "FeatureFlag"},
			{ID: "jetbrains.youtrack.feature.showWhosTyping", Enabled: true, Type: "FeatureFlag"},
			{ID: "jetbrains.youtrack.feature.linksWithSummary", Enabled: true, Type: "FeatureFlag"},
			{ID: "jetbrains.youtrack.feature.showFeaturesPromo", Enabled: true, Type: "FeatureFlag"},
			{ID: "jetbrains.youtrack.feature.excludeSuspendedUserRestrictions", Enabled: true, Type: "FeatureFlag"},
			{ID: "jetbrains.youtrack.feature.inlineComments", Enabled: true, Type: "FeatureFlag"},
			{ID: "jetbrains.youtrack.maintenance.prodMetricsBackend", Enabled: true, Type: "FeatureFlag"},
			{ID: "jetbrains.youtrack.feature.textRecognition", Enabled: true, Type: "FeatureFlag"},
			{ID: "jetbrains.youtrack.feature.prod", Enabled: true, Type: "FeatureFlag"},
			{ID: "jetbrains.youtrack.feature.enableMentions", Enabled: true, Type: "FeatureFlag"},
			{ID: "jetbrains.youtrack.feature.sentry", Enabled: true, Type: "FeatureFlag"},
			{ID: "jetbrains.youtrack.feature.showExternalIssueField", Enabled: true, Type: "FeatureFlag"},
			{ID: "jetbrains.youtrack.feature.documentFeatures", Enabled: true, Type: "FeatureFlag"},
			{ID: "jetbrains.youtrack.feature.anonymousDraft", Enabled: true, Type: "FeatureFlag"},
			{ID: "jetbrains.youtrack.feature.searchABExperiment", Enabled: true, Type: "FeatureFlag"},
			{ID: "jetbrains.youtrack.feature.showOnboardingQuestionnaire", Enabled: true, Type: "FeatureFlag"},
			{ID: "jetbrains.youtrack.feature.onboardingSuggestDemo", Enabled: true, Type: "FeatureFlag"},
			{ID: "jetbrains.youtrack.feature.contextTextCompletion", Enabled: true, Type: "FeatureFlag"},
			{ID: "jetbrains.youtrack.feature.enableGoogleAnalytics", Enabled: true, Type: "FeatureFlag"},
			{ID: "jetbrains.jetpass.auth.promoteOnLogin", Enabled: true, Type: "FeatureFlag"},
			{ID: "jetbrains.youtrack.feature.onboardingTour", Enabled: true, Type: "FeatureFlag"},
			{ID: "jetbrains.youtrack.feature.showImageNodesOnWhiteboards", Enabled: true, Type: "FeatureFlag"},
			{ID: "jetbrains.youtrack.feature.sendStatisticsToAmplitude", Enabled: true, Type: "FeatureFlag"},
			{ID: "jetbrains.youtrack.feature.semanticSearchIndexing", Enabled: true, Type: "FeatureFlag"},
		}
	}

	if user.Widgets == nil {
		user.Widgets = []*domain.Widget{}
	}

	user.Type = "Me"
	return user, nil
}

// GetCurrentUserDetail يعيد مخطط المستخدم الحالي من قاعدة البيانات وفقاً للأقسام
// المطلوبة فقط (sel nil = كل شيء) — (profiles, widgets, featureFlags, userType,
// issueRelatedGroup) دون بيانات افتراضية.
func (uc *UserUseCase) GetCurrentUserDetail(userID string, sel *domain.UserFieldSelect) (*domain.CurrentUser, error) {
	if uc.userRepo == nil {
		return nil, nil
	}
	return uc.userRepo.GetCurrentUserDetail(userID, sel)
}

func (uc *UserUseCase) GetGrazieProfile() domain.GrazieUserProfile {
	return domain.GrazieUserProfile{
		ExcludedIssueTypes:            "",
		EnableSpellChecker:            true,
		HasMoreTokens:                 true,
		EnableTextCompletion:          true,
		CycleRestart:                  1787653099967,
		FreeLicense:                   false,
		SpellCheckerEnabledInSystem:   true,
		TextCompletionEnabledInSystem: true,
		Enabled:                       true,
		Type:                          "GrazieUserProfile",
	}
}

func (uc *UserUseCase) GetUserByID(userID string) (*domain.User, error) {
	if uc.userRepo != nil {
		user, err := uc.userRepo.GetByID(userID)
		if err == nil && user != nil {
			return user, nil
		}
	}
	return nil, nil
}

func (uc *UserUseCase) GetAllUsers() ([]*domain.User, error) {
	if uc.userRepo != nil {
		if lister, ok := uc.userRepo.(interface {
			List() ([]*domain.User, error)
		}); ok {
			users, err := lister.List()
			if err == nil && users != nil {
				return users, nil
			}
		}
	}
	return []*domain.User{}, nil
}
