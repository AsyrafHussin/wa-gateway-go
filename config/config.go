package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	Port     int
	Host     string
	APIKey   string
	LogLevel string

	// CORS
	CORSOrigins string

	// Phone validation
	PhoneCountryCode string
	PhoneMinLength   int
	PhoneMaxLength   int

	// WhatsApp
	DataDir         string
	TypingDelay     int
	AutoReadReceipt bool

	// Webhook
	WebhookURL     string
	WebhookSecret  string
	WebhookTimeout int

	// Rate limiting
	RateLimitDevices  int
	RateLimitMessages int
	RateLimitValidate int

	// Cache
	CacheTTL int
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Port:              getEnvInt("PORT", 4010),
		Host:              getEnv("HOST", "0.0.0.0"),
		APIKey:            getEnv("API_KEY", ""),
		LogLevel:          getEnv("LOG_LEVEL", "info"),
		CORSOrigins:       getEnv("CORS_ORIGINS", "*"),
		PhoneCountryCode:  getEnv("PHONE_COUNTRY_CODE", "60"),
		PhoneMinLength:    getEnvInt("PHONE_MIN_LENGTH", 11),
		PhoneMaxLength:    getEnvInt("PHONE_MAX_LENGTH", 12),
		DataDir:           getEnv("DATA_DIR", "./data"),
		TypingDelay:       getEnvInt("TYPING_DELAY_MS", 1000),
		AutoReadReceipt:   getEnvBool("AUTO_READ_RECEIPT", false),
		WebhookURL:        getEnv("WEBHOOK_URL", ""),
		WebhookSecret:     getEnv("WEBHOOK_SECRET", ""),
		WebhookTimeout:    getEnvInt("WEBHOOK_TIMEOUT_MS", 5000),
		RateLimitDevices:  getEnvInt("RATE_LIMIT_DEVICES", 10),
		RateLimitMessages: getEnvInt("RATE_LIMIT_MESSAGES", 30),
		RateLimitValidate: getEnvInt("RATE_LIMIT_VALIDATE", 60),
		CacheTTL:          getEnvInt("CACHE_TTL_SECONDS", 3600),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.APIKey == "" {
		return fmt.Errorf("API_KEY is required")
	}
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("PORT must be between 1 and 65535")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}
