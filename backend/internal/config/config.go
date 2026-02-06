package config

import (
	"log"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Env           string
	Port          string
	DatabaseDSN   string
	SessionSecret string
	CookieSecure  bool
	UploadDir     string
	MaxUploadMB   int64
	DBMaxOpen     int
	DBMaxIdle     int
	DBMaxLifetime int
	AllowedTypes  []string
	RateLimitMax  int
	RateLimitTTL  int
}

func Load() Config {
	return Config{
		Env:           getEnv("APP_ENV", "development"),
		Port:          getEnv("PORT", "8080"),
		DatabaseDSN:   getEnv("MYSQL_DSN", ""),
		SessionSecret: getEnv("SESSION_SECRET", ""),
		CookieSecure:  getEnv("COOKIE_SECURE", "false") == "true",
		UploadDir:     getEnv("UPLOAD_DIR", "./uploads"),
		MaxUploadMB:   getEnvInt("MAX_UPLOAD_MB", 10),
		DBMaxOpen:     int(getEnvInt("DB_MAX_OPEN", 10)),
		DBMaxIdle:     int(getEnvInt("DB_MAX_IDLE", 5)),
		DBMaxLifetime: int(getEnvInt("DB_MAX_LIFETIME_MINUTES", 5)),
		AllowedTypes:  getEnvList("ALLOWED_UPLOAD_TYPES", []string{"image/jpeg", "image/png", "application/pdf"}),
		RateLimitMax:  int(getEnvInt("RATE_LIMIT_MAX", 200)),
		RateLimitTTL:  int(getEnvInt("RATE_LIMIT_WINDOW_SECONDS", 60)),
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
		if parsed, err := strconv.ParseInt(value, 10, 64); err == nil {
			return parsed
		}
		log.Printf("warning: could not parse env var %s=%q as integer, using fallback %d", key, value, fallback)
	}
	return fallback
}

func getEnvList(key string, fallback []string) []string {
	if value := os.Getenv(key); value != "" {
		parts := strings.Split(value, ",")
		var cleaned []string
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				cleaned = append(cleaned, trimmed)
			}
		}
		if len(cleaned) > 0 {
			return cleaned
		}
	}
	return fallback
}
