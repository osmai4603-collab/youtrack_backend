package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"youtrack_backend/internal/domain"
	"youtrack_backend/internal/usecase"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	userUC    *usecase.UserUseCase
	jwtSecret string
}

func NewAuthHandler(userUC *usecase.UserUseCase, jwtSecret string) *AuthHandler {
	if jwtSecret == "" {
		jwtSecret = "default-youtrack-jwt-secret-key-change-in-prod"
	}
	return &AuthHandler{
		userUC:    userUC,
		jwtSecret: jwtSecret,
	}
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Login    string `json:"login"`
	Email    string `json:"email"`
	FullName string `json:"fullName"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token     string       `json:"token"`
	TokenType string       `json:"token_type"`
	ExpiresIn int64        `json:"expires_in"`
	User      *domain.User `json:"user"`
}

// Login يعالج تسجيل الدخول وإصدار رمز JWT
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Login == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Login and password are required"})
		return
	}

	user, err := h.userUC.GetCurrentUser(req.Login)
	if err != nil || user == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid login credentials"})
		return
	}

	// توليد JWT Token
	claims := jwt.MapClaims{
		"sub":   user.ID,
		"login": user.Login,
		"exp":   time.Now().Add(72 * time.Hour).Unix(),
		"iat":   time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Could not generate authentication token"})
		return
	}

	resp := AuthResponse{
		Token:     tokenString,
		TokenType: "Bearer",
		ExpiresIn: 72 * 3600,
		User:      user,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// Register يعالج إنشاء حساب مستخدم جديد
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Login == "" || req.Password == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Login, password, and email are required"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Error processing password"})
		return
	}
	_ = hashedPassword

	newUser := &domain.User{
		ID:            req.Login,
		Login:         req.Login,
		Email:         req.Email,
		FullName:      req.FullName,
		Name:          req.Login,
		LocalizedName: req.FullName,
		AvatarURL:     "/hub/api/rest/avatar/" + req.Login,
		Online:        true,
		UserType: &domain.UserType{
			ID:   "STANDARD",
			Name: "Standard",
			Type: "UserType",
		},
		Type: "User",
	}

	claims := jwt.MapClaims{
		"sub":   newUser.ID,
		"login": newUser.Login,
		"exp":   time.Now().Add(72 * time.Hour).Unix(),
		"iat":   time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(h.jwtSecret))

	resp := AuthResponse{
		Token:     tokenString,
		TokenType: "Bearer",
		ExpiresIn: 72 * 3600,
		User:      newUser,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}
