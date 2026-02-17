package config

import (
	"os"
	"testing"
)

func clearConfigEnv() {
	envVars := []string{
		"API_KEY", "PORT", "HOST", "LOG_LEVEL", "CORS_ORIGINS",
		"PHONE_COUNTRY_CODE", "PHONE_MIN_LENGTH", "PHONE_MAX_LENGTH",
		"DATA_DIR", "TYPING_DELAY_MS", "AUTO_READ_RECEIPT",
		"WEBHOOK_URL", "WEBHOOK_SECRET", "WEBHOOK_TIMEOUT_MS",
		"RATE_LIMIT_DEVICES", "RATE_LIMIT_MESSAGES", "RATE_LIMIT_VALIDATE",
		"CACHE_TTL_SECONDS",
		"WS_ALLOWED_ORIGINS", "WS_AUTH_TIMEOUT",
	}
	for _, v := range envVars {
		_ = os.Unsetenv(v)
	}
}

func TestLoad_RequiresAPIKey(t *testing.T) {
	clearConfigEnv()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing API_KEY")
	}
	if err.Error() != "API_KEY is required" {
		t.Errorf("expected 'API_KEY is required', got %q", err.Error())
	}
}

func TestLoad_Defaults(t *testing.T) {
	clearConfigEnv()
	t.Setenv("API_KEY", "test-key-123")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != 4010 {
		t.Errorf("expected Port 4010, got %d", cfg.Port)
	}
	if cfg.Host != "0.0.0.0" {
		t.Errorf("expected Host '0.0.0.0', got %q", cfg.Host)
	}
	if cfg.APIKey != "test-key-123" {
		t.Errorf("expected APIKey 'test-key-123', got %q", cfg.APIKey)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("expected LogLevel 'info', got %q", cfg.LogLevel)
	}
	if cfg.CORSOrigins != "*" {
		t.Errorf("expected CORSOrigins '*', got %q", cfg.CORSOrigins)
	}
	if cfg.PhoneCountryCode != "60" {
		t.Errorf("expected PhoneCountryCode '60', got %q", cfg.PhoneCountryCode)
	}
	if cfg.PhoneMinLength != 11 {
		t.Errorf("expected PhoneMinLength 11, got %d", cfg.PhoneMinLength)
	}
	if cfg.PhoneMaxLength != 12 {
		t.Errorf("expected PhoneMaxLength 12, got %d", cfg.PhoneMaxLength)
	}
	if cfg.DataDir != "./data" {
		t.Errorf("expected DataDir './data', got %q", cfg.DataDir)
	}
	if cfg.TypingDelay != 1000 {
		t.Errorf("expected TypingDelay 1000, got %d", cfg.TypingDelay)
	}
	if cfg.AutoReadReceipt != false {
		t.Errorf("expected AutoReadReceipt false, got %v", cfg.AutoReadReceipt)
	}
	if cfg.WebhookURL != "" {
		t.Errorf("expected WebhookURL '', got %q", cfg.WebhookURL)
	}
	if cfg.WebhookSecret != "" {
		t.Errorf("expected WebhookSecret '', got %q", cfg.WebhookSecret)
	}
	if cfg.WebhookTimeout != 5000 {
		t.Errorf("expected WebhookTimeout 5000, got %d", cfg.WebhookTimeout)
	}
	if cfg.RateLimitDevices != 10 {
		t.Errorf("expected RateLimitDevices 10, got %d", cfg.RateLimitDevices)
	}
	if cfg.RateLimitMessages != 30 {
		t.Errorf("expected RateLimitMessages 30, got %d", cfg.RateLimitMessages)
	}
	if cfg.RateLimitValidate != 60 {
		t.Errorf("expected RateLimitValidate 60, got %d", cfg.RateLimitValidate)
	}
	if cfg.CacheTTL != 3600 {
		t.Errorf("expected CacheTTL 3600, got %d", cfg.CacheTTL)
	}
	if cfg.WSAllowedOrigins != "*" {
		t.Errorf("expected WSAllowedOrigins '*', got %q", cfg.WSAllowedOrigins)
	}
	if cfg.WSAuthTimeout != 5 {
		t.Errorf("expected WSAuthTimeout 5, got %d", cfg.WSAuthTimeout)
	}
}

func TestLoad_CustomValues(t *testing.T) {
	clearConfigEnv()
	t.Setenv("API_KEY", "my-secret-key")
	t.Setenv("PORT", "8080")
	t.Setenv("HOST", "127.0.0.1")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("CORS_ORIGINS", "http://localhost:3000")
	t.Setenv("PHONE_COUNTRY_CODE", "65")
	t.Setenv("PHONE_MIN_LENGTH", "10")
	t.Setenv("PHONE_MAX_LENGTH", "10")
	t.Setenv("DATA_DIR", "/tmp/wa-data")
	t.Setenv("TYPING_DELAY_MS", "500")
	t.Setenv("AUTO_READ_RECEIPT", "true")
	t.Setenv("WEBHOOK_URL", "https://example.com/webhook")
	t.Setenv("WEBHOOK_SECRET", "webhook-secret")
	t.Setenv("WEBHOOK_TIMEOUT_MS", "10000")
	t.Setenv("RATE_LIMIT_DEVICES", "5")
	t.Setenv("RATE_LIMIT_MESSAGES", "60")
	t.Setenv("RATE_LIMIT_VALIDATE", "120")
	t.Setenv("CACHE_TTL_SECONDS", "7200")
	t.Setenv("WS_ALLOWED_ORIGINS", "http://localhost:3000,https://app.example.com")
	t.Setenv("WS_AUTH_TIMEOUT", "10")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != 8080 {
		t.Errorf("expected Port 8080, got %d", cfg.Port)
	}
	if cfg.Host != "127.0.0.1" {
		t.Errorf("expected Host '127.0.0.1', got %q", cfg.Host)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("expected LogLevel 'debug', got %q", cfg.LogLevel)
	}
	if cfg.CORSOrigins != "http://localhost:3000" {
		t.Errorf("expected CORSOrigins 'http://localhost:3000', got %q", cfg.CORSOrigins)
	}
	if cfg.PhoneCountryCode != "65" {
		t.Errorf("expected PhoneCountryCode '65', got %q", cfg.PhoneCountryCode)
	}
	if cfg.PhoneMinLength != 10 {
		t.Errorf("expected PhoneMinLength 10, got %d", cfg.PhoneMinLength)
	}
	if cfg.PhoneMaxLength != 10 {
		t.Errorf("expected PhoneMaxLength 10, got %d", cfg.PhoneMaxLength)
	}
	if cfg.DataDir != "/tmp/wa-data" {
		t.Errorf("expected DataDir '/tmp/wa-data', got %q", cfg.DataDir)
	}
	if cfg.TypingDelay != 500 {
		t.Errorf("expected TypingDelay 500, got %d", cfg.TypingDelay)
	}
	if cfg.AutoReadReceipt != true {
		t.Errorf("expected AutoReadReceipt true, got %v", cfg.AutoReadReceipt)
	}
	if cfg.WebhookURL != "https://example.com/webhook" {
		t.Errorf("expected WebhookURL 'https://example.com/webhook', got %q", cfg.WebhookURL)
	}
	if cfg.WebhookSecret != "webhook-secret" {
		t.Errorf("expected WebhookSecret 'webhook-secret', got %q", cfg.WebhookSecret)
	}
	if cfg.WebhookTimeout != 10000 {
		t.Errorf("expected WebhookTimeout 10000, got %d", cfg.WebhookTimeout)
	}
	if cfg.RateLimitDevices != 5 {
		t.Errorf("expected RateLimitDevices 5, got %d", cfg.RateLimitDevices)
	}
	if cfg.RateLimitMessages != 60 {
		t.Errorf("expected RateLimitMessages 60, got %d", cfg.RateLimitMessages)
	}
	if cfg.RateLimitValidate != 120 {
		t.Errorf("expected RateLimitValidate 120, got %d", cfg.RateLimitValidate)
	}
	if cfg.CacheTTL != 7200 {
		t.Errorf("expected CacheTTL 7200, got %d", cfg.CacheTTL)
	}
	if cfg.WSAllowedOrigins != "http://localhost:3000,https://app.example.com" {
		t.Errorf("expected WSAllowedOrigins 'http://localhost:3000,https://app.example.com', got %q", cfg.WSAllowedOrigins)
	}
	if cfg.WSAuthTimeout != 10 {
		t.Errorf("expected WSAuthTimeout 10, got %d", cfg.WSAuthTimeout)
	}
}

func TestLoad_InvalidPort(t *testing.T) {
	clearConfigEnv()
	t.Setenv("API_KEY", "test-key")
	t.Setenv("PORT", "99999")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid port")
	}
}

func TestLoad_ZeroPort(t *testing.T) {
	clearConfigEnv()
	t.Setenv("API_KEY", "test-key")
	t.Setenv("PORT", "0")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for port 0")
	}
}

func TestLoad_NegativePort(t *testing.T) {
	clearConfigEnv()
	t.Setenv("API_KEY", "test-key")
	t.Setenv("PORT", "-1")

	// Negative port: Atoi returns -1, which fails validation
	_, err := Load()
	if err == nil {
		t.Fatal("expected error for negative port")
	}
}

func TestLoad_WSAuthTimeoutZero(t *testing.T) {
	clearConfigEnv()
	t.Setenv("API_KEY", "test-key")
	t.Setenv("WS_AUTH_TIMEOUT", "0")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for WS_AUTH_TIMEOUT 0")
	}
}

func TestLoad_WSAuthTimeoutTooHigh(t *testing.T) {
	clearConfigEnv()
	t.Setenv("API_KEY", "test-key")
	t.Setenv("WS_AUTH_TIMEOUT", "61")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for WS_AUTH_TIMEOUT > 60")
	}
}

func TestLoad_NonNumericPort(t *testing.T) {
	clearConfigEnv()
	t.Setenv("API_KEY", "test-key")
	t.Setenv("PORT", "abc")

	// Non-numeric port falls back to default 4010, which is valid
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 4010 {
		t.Errorf("expected fallback Port 4010, got %d", cfg.Port)
	}
}

func TestLoad_NonNumericBool(t *testing.T) {
	clearConfigEnv()
	t.Setenv("API_KEY", "test-key")
	t.Setenv("AUTO_READ_RECEIPT", "not-a-bool")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Falls back to default false
	if cfg.AutoReadReceipt != false {
		t.Errorf("expected AutoReadReceipt false for invalid value, got %v", cfg.AutoReadReceipt)
	}
}

func TestGetEnv(t *testing.T) {
	t.Setenv("TEST_GET_ENV_KEY", "value123")

	if got := getEnv("TEST_GET_ENV_KEY", "fallback"); got != "value123" {
		t.Errorf("expected 'value123', got %q", got)
	}
	if got := getEnv("TEST_GET_ENV_MISSING", "fallback"); got != "fallback" {
		t.Errorf("expected 'fallback', got %q", got)
	}
}

func TestGetEnvInt(t *testing.T) {
	t.Setenv("TEST_INT_KEY", "42")

	if got := getEnvInt("TEST_INT_KEY", 0); got != 42 {
		t.Errorf("expected 42, got %d", got)
	}
	if got := getEnvInt("TEST_INT_MISSING", 99); got != 99 {
		t.Errorf("expected fallback 99, got %d", got)
	}

	t.Setenv("TEST_INT_INVALID", "not-a-number")

	if got := getEnvInt("TEST_INT_INVALID", 77); got != 77 {
		t.Errorf("expected fallback 77, got %d", got)
	}
}

func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		value    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1", true},
		{"0", false},
		{"TRUE", true},
		{"FALSE", false},
	}

	for _, tt := range tests {
		t.Setenv("TEST_BOOL_KEY", tt.value)
		got := getEnvBool("TEST_BOOL_KEY", !tt.expected)
		if got != tt.expected {
			t.Errorf("getEnvBool(%q) = %v, expected %v", tt.value, got, tt.expected)
		}
	}
	_ = os.Unsetenv("TEST_BOOL_KEY")

	// Missing key should return fallback
	if got := getEnvBool("TEST_BOOL_MISSING", true); got != true {
		t.Errorf("expected fallback true, got %v", got)
	}
}
