package config

import (
	"fmt"
	"os"
)

type Config struct {
	Env           string
	Port          string
	DatabaseDSN   string
	SessionSecret string
	CookieSecure  bool
	UploadDir     string
	MaxUploadMB   int64
}

func Load() Config {
	return Config{
		Env:           getEnv("APP_ENV", "development"),
		Port:          getEnv("PORT", "8080"),
		DatabaseDSN:   getEnv("MYSQL_DSN", "root:password@tcp(mysql:3306)/familyjournal?parseTime=true"),
		SessionSecret: getEnv("SESSION_SECRET", "super-secret"),
		CookieSecure:  getEnv("COOKIE_SECURE", "false") == "true",
		UploadDir:     getEnv("UPLOAD_DIR", "./uploads"),
		MaxUploadMB:   getEnvInt("MAX_UPLOAD_MB", 10),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int64) int64 {
	if value := os.Getenv(key); value != "" {
		var parsed int64
		_, err := fmt.Sscanf(value, "%d", &parsed)
		if err == nil {
			return parsed
		}
	}
	return fallback
}
