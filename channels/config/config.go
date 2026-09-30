package config

import (
	"os"
	"youtrack_backend/channels/model"
)

// Load يقرأ إعدادات الخادم من متغيرات البيئة ويطبّق القيم الافتراضية.
// مستوحى من أسلوب Mattermost في استخدام متغيرات البيئة، لكن دون الاعتماد على Viper.
func Load() *model.ServerConfig {
	cfg := &model.ServerConfig{
		ServerPort: getEnv("SERVER_PORT", "8099"),
		AppEnv:     getEnv("APP_ENV", "development"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "8090"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "secret"),
		DBName:     getEnv("DB_NAME", "youtrack_db"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
		JWTSecret:  getEnv("JWT_SECRET", "default-youtrack-jwt-secret-key-change-in-prod"),
	}
	return cfg
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
