package middleware

import (
	"crypto/subtle"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/AsyrafHussin/wa-gateway-go/pkg/response"
)

type Auth struct {
	apiKey []byte
}

func NewAuth(apiKey string) *Auth {
	return &Auth{apiKey: []byte(apiKey)}
}

func (a *Auth) Require() fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := extractAPIKey(c)
		if key == "" {
			return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Missing API key")
		}

		if subtle.ConstantTimeCompare([]byte(key), a.apiKey) != 1 {
			return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Invalid API key")
		}

		return c.Next()
	}
}

func extractAPIKey(c *fiber.Ctx) string {
	// Check Authorization: Bearer <key>
	if auth := c.Get("Authorization"); auth != "" {
		if strings.HasPrefix(auth, "Bearer ") {
			return strings.TrimPrefix(auth, "Bearer ")
		}
	}

	// Check X-API-Key header
	if key := c.Get("X-API-Key"); key != "" {
		return key
	}

	return ""
}
