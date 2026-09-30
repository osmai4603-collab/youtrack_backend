package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"youtrack_backend/channels/app"
	"youtrack_backend/channels/model"
	"youtrack_backend/channels/request"
)

// LoginRequest يمثّل مخطط طلب تسجيل الدخول.
type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// RegisterRequest يمثّل مخطط طلب إنشاء حساب.
type RegisterRequest struct {
	Login    string `json:"login"`
	Email    string `json:"email"`
	FullName string `json:"fullName"`
	Password string `json:"password"`
}

// AuthResponse يمثّل مخطط استجابة المصادقة.
type AuthResponse struct {
	Token     string      `json:"token"`
	TokenType string      `json:"token_type"`
	ExpiresIn int64       `json:"expires_in"`
	User      *model.User `json:"user"`
}

// AuthHandler يعالج عمليات المصادقة.
type AuthHandler struct {
	app       *app.YouTrackApp
	jwtSecret string
}

func NewAuthHandler(a *app.YouTrackApp, jwtSecret string) *AuthHandler {
	return &AuthHandler{app: a, jwtSecret: jwtSecret}
}

func (api *API) InitAuth() {
	handler := NewAuthHandler(api.newApp(), api.srv.JWTSecret())
	api.BaseRoutes.Auth.Method("POST", "/login", api.APIHandler(func(c *Context, w http.ResponseWriter, r *http.Request) {
		handler.Login(w, r)
	}))

	api.BaseRoutes.Auth.Method("POST", "/register", api.APIHandler(func(c *Context, w http.ResponseWriter, r *http.Request) {
		handler.Register(w, r)
	}))
}

func (h *AuthHandler) signToken(user *model.User) (string, error) {
	roles := user.Roles
	if len(roles) == 0 {
		roles = []string{"user"}
	}

	claims := jwt.MapClaims{
		"sub":   user.ID,
		"login": user.Login,
		"roles": roles,
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(72 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := h.jwtSecret
	if secret == "" && h.app != nil {
		secret = h.app.JWTSecret()
	}
	if secret == "" {
		secret = "secret"
	}
	return token.SignedString([]byte(secret))
}

const tokenTTL = 72 * 3600

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, model.NewBadRequestError("Auth.Login", "invalid request body"))
		return
	}

	appCtx := request.NewContext(r.Context(), r.Header.Get("X-Request-ID"), r.RemoteAddr, r.URL.Path, r.UserAgent())
	user, err := h.app.Authenticate(appCtx, req.Login, req.Password)
	if err != nil {
		writeError(w, err)
		return
	}

	token, signErr := h.signToken(user)
	if signErr != nil {
		writeError(w, model.NewInternalError("Auth.Login", "could not generate token", signErr))
		return
	}

	writeJSON(w, http.StatusOK, AuthResponse{Token: token, TokenType: "Bearer", ExpiresIn: tokenTTL, User: user})
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, model.NewBadRequestError("Auth.Register", "invalid request body"))
		return
	}

	appCtx := request.NewContext(r.Context(), r.Header.Get("X-Request-ID"), r.RemoteAddr, r.URL.Path, r.UserAgent())
	user, err := h.app.Register(appCtx, app.RegisterCredentials{
		Login:    req.Login,
		Email:    req.Email,
		FullName: req.FullName,
		Password: req.Password,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	token, signErr := h.signToken(user)
	if signErr != nil {
		writeError(w, model.NewInternalError("Auth.Register", "could not generate token", signErr))
		return
	}

	writeJSON(w, http.StatusCreated, AuthResponse{Token: token, TokenType: "Bearer", ExpiresIn: tokenTTL, User: user})
}
