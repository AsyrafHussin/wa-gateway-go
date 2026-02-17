package middleware

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/AsyrafHussin/wa-gateway-go/pkg/response"
)

type window struct {
	count   int
	resetAt time.Time
}

type rateLimiter struct {
	mu      sync.Mutex
	windows map[string]*window
	limit   int
	window  time.Duration
}

func RateLimit(requestsPerMinute int) fiber.Handler {
	rl := &rateLimiter{
		windows: make(map[string]*window),
		limit:   requestsPerMinute,
		window:  time.Minute,
	}

	// Cleanup expired entries periodically
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			rl.cleanup()
		}
	}()

	return func(c *fiber.Ctx) error {
		key := c.IP() + ":" + c.Path()

		rl.mu.Lock()
		w, exists := rl.windows[key]
		now := time.Now()

		if !exists || now.After(w.resetAt) {
			w = &window{count: 0, resetAt: now.Add(rl.window)}
			rl.windows[key] = w
		}

		w.count++
		count := w.count
		rl.mu.Unlock()

		if count > rl.limit {
			return response.Error(c, fiber.StatusTooManyRequests, "RATE_LIMITED", "Too many requests, please try again later")
		}

		return c.Next()
	}
}

func (rl *rateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	for key, w := range rl.windows {
		if now.After(w.resetAt) {
			delete(rl.windows, key)
		}
	}
}
