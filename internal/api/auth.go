package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"youtrack_backend/internal/app"
	"youtrack_backend/internal/model"
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

// AuthResponse يمثّل مخطط استجابة المصادقة (مشترك بين login و register).
type AuthResponse struct {
	Token     string      `json:"token"`
	TokenType string      `json:"token_type"`
	ExpiresIn int64       `json:"expires_in"`
	User      *model.User `json:"user"`
}

// AuthHandler يعالج عمليات المصادقة.
type AuthHandler struct {
	app       *app.App
	jwtSecret string
}

func NewAuthHandler(a *app.App, jwtSecret string) *AuthHandler {
	return &AuthHandler{app: a, jwtSecret: jwtSecret}
}

// signToken يولّد JWT للمستخدم.
func (h *AuthHandler) signToken(user *model.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":   user.ID,
		"login": user.Login,
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(72 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.jwtSecret))
}

const tokenTTL = 72 * 3600

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, model.BadRequest("invalid request body"))
		return
	}

	user, err := h.app.Authenticate(r.Context(), req.Login, req.Password)
	if err != nil {
		writeError(w, err)
		return
	}

	token, err := h.signToken(user)
	if err != nil {
		writeError(w, model.Internal("could not generate token"))
		return
	}

	writeJSON(w, http.StatusOK, AuthResponse{Token: token, TokenType: "Bearer", ExpiresIn: tokenTTL, User: user})
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, model.BadRequest("invalid request body"))
		return
	}

	user, err := h.app.Register(r.Context(), app.RegisterCredentials{
		Login:    req.Login,
		Email:    req.Email,
		FullName: req.FullName,
		Password: req.Password,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	token, err := h.signToken(user)
	if err != nil {
		writeError(w, model.Internal("could not generate token"))
		return
	}

	writeJSON(w, http.StatusCreated, AuthResponse{Token: token, TokenType: "Bearer", ExpiresIn: tokenTTL, User: user})
}
