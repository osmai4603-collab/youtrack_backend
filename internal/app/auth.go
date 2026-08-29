package app

import (
	"context"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"youtrack_backend/internal/model"
)

// RegisterCredentials يحمل بيانات إنشاء حساب جديد.
type RegisterCredentials struct {
	Login    string
	Email    string
	FullName string
	Password string
}

// Register ينشئ حساب مستخدم جديد بتشفير كلمة المرور.
func (a *App) Register(ctx context.Context, cred RegisterCredentials) (*model.User, error) {
	login := strings.TrimSpace(cred.Login)
	if login == "" || cred.Password == "" {
		return nil, model.BadRequest("login and password are required")
	}

	if _, err := a.store.Users().GetByLogin(ctx, login); err == nil {
		return nil, model.BadRequest("login %q already exists", login)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cred.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, model.Internal("failed to hash password")
	}

	u := &model.User{
		ID:              login,
		Login:           login,
		Email:           cred.Email,
		FullName:        cred.FullName,
		Name:            login,
		UserTypeID:      "STANDARD_USER",
		UserType:        &model.UserType{ID: "STANDARD_USER", Name: "Standard user"},
		Online:          true,
		CanReadProfile:  true,
		IsEmailVerified: false,
		PasswordHash:    string(hash),
	}

	if err := a.store.Users().Create(ctx, u); err != nil {
		return nil, model.Internal("failed to create user")
	}
	u.Normalize()
	return u, nil
}

// Authenticate يتحقق من صحة بيانات الدخول ويعيد المستخدم إن صحّت.
func (a *App) Authenticate(ctx context.Context, login, password string) (*model.User, error) {
	if login == "" || password == "" {
		return nil, model.BadRequest("login and password are required")
	}

	u, err := a.store.Users().GetByLogin(ctx, login)
	if err != nil {
		return nil, model.Unauthorized("invalid credentials")
	}
	if u.PasswordHash == "" {
		return nil, model.Unauthorized("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, model.Unauthorized("invalid credentials")
	}
	u.PasswordHash = ""
	u.Normalize()
	return u, nil
}
