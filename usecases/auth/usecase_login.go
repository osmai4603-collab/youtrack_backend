package authusecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"youtrack_backend/internal/domain"
	perms2 "youtrack_backend/internal/services/permissions_services"
	authstorage "youtrack_backend/internal/storage/auth"
	storageusers "youtrack_backend/internal/storage/users"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrAccountLocked          = errors.New("account locked")
	ErrAccountInactive        = errors.New("account inactive")
	ErrPasswordChangeRequired = errors.New("password change required")
	ErrLoginRateLimited       = errors.New("login rate limited")
	ErrInvalidRefreshToken    = errors.New("invalid or expired session")
	ErrLogoutFailed           = errors.New("logout failed")
)

const (
	maxFailedLogins   = 5
	lockDuration      = 15 * time.Minute
	defaultSessionTTL = 24 * time.Hour
)

type UserFinder interface {
	FindByLogin(ctx context.Context, login string) (*domain.User, error)
}

type Hello struct {
	UserFinder
}

// AuthDependencies abstracts the user store operations needed for authentication.
type AuthDependencies interface {
	UserFinder
	FindByID(ctx context.Context, id string) (*domain.User, error)
	UpdateLoginState(ctx context.Context, userID string, state domain.LoginState) error
}

// RoleFinder resolves the current role IDs of a user in a given scope.
type RoleFinder interface {
	GetRolesByUserID(ctx context.Context, userID string, scope *perms2.ScopeContext) ([]string, error)
}

// AuthLoginUseCase authenticates credentials and creates a server-side session.
type AuthLoginUseCase struct {
	users      AuthDependencies
	sessions   authstorage.SessionRepository
	roles      RoleFinder
	logger     *slog.Logger
	sessionTTL time.Duration
}

// NewLogin builds the login use case.
func NewLogin(logger *slog.Logger, users AuthDependencies, sessions authstorage.SessionRepository, roles RoleFinder, sessionTTL time.Duration) *AuthLoginUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	if sessionTTL <= 0 {
		sessionTTL = defaultSessionTTL
	}

	return &AuthLoginUseCase{
		users:      users,
		sessions:   sessions,
		roles:      roles,
		logger:     logger,
		sessionTTL: sessionTTL,
	}
}

// LoginParams carries the credentials supplied to the login endpoint.
type LoginParams struct {
	Username string
	Password string
}

// LoginResult is returned by Login and Refresh and carries the user plus an active session.
type LoginResult struct {
	User    *domain.User
	Session *domain.UserSession
	Roles   []string
}

// RandomSessionID generates a cryptographically secure 32-byte hex string.
func RandomSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate random session id: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// TryLogin authenticates the credentials and creates a new session.
func (uc *AuthLoginUseCase) TryLogin(ctx context.Context, params *LoginParams) (*LoginResult, error) {
	if params == nil || params.Username == "" || params.Password == "" {
		return nil, ErrInvalidCredentials
	}

	user, err := uc.Authenticate(ctx, params.Username, params.Password)
	if err != nil {
		return nil, err
	}

	roles := resolveRoles(ctx, uc.roles, uc.logger, user.ID)

	sessionID, err := RandomSessionID()
	if err != nil {
		return nil, fmt.Errorf("generate session id: %w", err)
	}

	now := time.Now().UTC()
	session := domain.UserSession{
		ID:        sessionID,
		SessionID: sessionID,
		UserID:    user.ID,
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: now.Add(uc.sessionTTL),
	}

	if uc.sessions != nil {
		if err := uc.sessions.Create(ctx, session); err != nil {
			return nil, fmt.Errorf("create session: %w", err)
		}
	}

	return &LoginResult{User: user, Session: &session, Roles: roles}, nil
}

// Authenticate validates user credentials and account status.
func (uc *AuthLoginUseCase) Authenticate(ctx context.Context, username string, password string) (*domain.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	user, err := uc.users.FindByLogin(ctx, username)
	if err != nil {
		if errors.Is(err, storageusers.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("error finding user for login: %w", err)
	}

	if user.PasswordHash == nil || *user.PasswordHash == "" {
		return nil, ErrInvalidCredentials
	}

	if bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password)) != nil {
		_ = uc.recordFailure(ctx, user, time.Now().UTC())
		return nil, ErrInvalidCredentials
	}

	if user.Banned {
		return nil, ErrAccountLocked
	}
	if user.IsLocked {
		if user.LockedUntil == nil || time.Now().UTC().Before(*user.LockedUntil) {
			return nil, ErrAccountInactive
		}
		// The lock has expired: lift it so the user can authenticate again.
		now := time.Now().UTC()
		state := domain.LoginState{LastLoginAt: &now, FailedLoginCount: 0}
		if err := uc.users.UpdateLoginState(ctx, user.ID, state); err != nil {
			return nil, fmt.Errorf("failed to lift expired account lock: %w", err)
		}
		user.IsLocked = false
		user.LockedUntil = nil
		user.FailedLoginCount = 0
	}
	if user.IsInactive() {
		return nil, ErrAccountInactive
	}

	now := time.Now().UTC()
	state := domain.LoginState{LastLoginAt: &now, FailedLoginCount: 0}
	if err := uc.users.UpdateLoginState(ctx, user.ID, state); err != nil {
		return nil, fmt.Errorf("failed to record successful login: %w", err)
	}

	user.FailedLoginCount = 0
	user.LastLoginAt = &now
	return user, nil
}

func (uc *AuthLoginUseCase) recordFailure(ctx context.Context, user *domain.User, now time.Time) error {
	failed := user.FailedLoginCount + 1
	state := domain.LoginState{
		FailedLoginCount:  failed,
		LastFailedLoginAt: &now,
	}
	if failed >= maxFailedLogins {
		lockedUntil := now.Add(lockDuration)
		state.LockedUntil = &lockedUntil
		state.IsLocked = true
	}
	if err := uc.users.UpdateLoginState(ctx, user.ID, state); err != nil {
		return fmt.Errorf("record failed login: %w", err)
	}
	return nil
}

// VerifyPassword compares a plaintext password against the user's bcrypt hash.
func VerifyPassword(user *domain.User, password string) error {
	if user == nil || user.PasswordHash == nil {
		return errors.New("user password hash is unavailable")
	}
	return CompareHash(password, *user.PasswordHash)
}

// CompareHash verifies a password against a bcrypt hash.
func CompareHash(password string, passwordHash string) error {
	return bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
}

// resolveRoles returns the user's current global roles, falling back to "user"
// when role resolution is unavailable.
func resolveRoles(ctx context.Context, finder RoleFinder, logger *slog.Logger, userID string) []string {
	if finder == nil {
		return []string{"user"}
	}
	roles, err := finder.GetRolesByUserID(ctx, userID, &perms2.ScopeContext{Type: perms2.ScopeGlobal})
	if err != nil {
		if logger != nil {
			logger.Warn("role resolution failed; falling back to default role",
				slog.String("user_id", userID),
				slog.Any("error", err),
			)
		}
		return []string{"user"}
	}
	if len(roles) == 0 {
		return []string{"user"}
	}
	return roles
}
