package app

import (
	"strings"

	"golang.org/x/crypto/bcrypt"

	"youtrack_backend/channels/model"
	"youtrack_backend/channels/request"
)

// RegisterCredentials يحمل بيانات إنشاء حساب جديد.
type RegisterCredentials struct {
	Login    string
	Email    string
	FullName string
	Password string
}

// Register ينشئ حساب مستخدم جديد بتشفير كلمة المرور.
func (a *YouTrackApp) Register(c request.CTX, cred RegisterCredentials) (*model.User, *model.AppError) {
	login := strings.TrimSpace(cred.Login)
	if login == "" || cred.Password == "" {
		return nil, model.NewBadRequestError("App.Register", "login and password are required")
	}
	ctx := c.Context()
	if _, err := a.Store().Users().GetByLogin(ctx, login); err == nil {
		return nil, model.NewBadRequestError("App.Register", "login already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cred.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, model.NewInternalError("App.Register", "failed to hash password", err)
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

	if err := a.Store().Users().Create(ctx, u); err != nil {
		return nil, model.NewInternalError("App.Register", "failed to create user", err)
	}
	u.Normalize()
	return u, nil
}

// Authenticate يتحقق من صحة بيانات الدخول ويعيد المستخدم إن صحّت.
func (a *YouTrackApp) Authenticate(c request.CTX, login, password string) (*model.User, *model.AppError) {
	if login == "" || password == "" {
		return nil, model.NewBadRequestError("App.Authenticate", "login and password are required")
	}
	ctx := c.Context()
	u, err := a.Store().Users().GetByLogin(ctx, login)
	if err != nil {
		return nil, model.NewUnauthorizedError("App.Authenticate", "invalid credentials")
	}
	if u.PasswordHash == "" {
		return nil, model.NewUnauthorizedError("App.Authenticate", "invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, model.NewUnauthorizedError("App.Authenticate", "invalid credentials")
	}
	u.PasswordHash = ""
	u.Normalize()
	return u, nil
}
