package middleware

import (
	"context"
	"fmt"
	"time"

	"refina-web-bff/internal/types/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

// RateLimitConfig holds the configuration for the rate limiter.
type RateLimitConfig struct {
	// Max number of requests allowed within the Window.
	Max int
	// Time window for the rate limit counter.
	Window time.Duration
}

// DefaultRateLimitConfig returns a sensible default: 60 requests per minute.
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Max:    60,
		Window: 1 * time.Minute,
	}
}

// RateLimiterMiddleware creates a Fiber middleware that enforces per-IP rate
// limiting using Redis. Each IP gets a sliding window counter stored in Redis.
func RateLimiterMiddleware(redisClient *redis.Client, cfg ...RateLimitConfig) fiber.Handler {
	config := DefaultRateLimitConfig()
	if len(cfg) > 0 {
		config = cfg[0]
	}

	return func(c *fiber.Ctx) error {
		ip := c.IP()
		key := fmt.Sprintf("rate_limit:%s", ip)
		ctx := context.Background()

		// Increment the counter for this IP
		count, err := redisClient.Incr(ctx, key).Result()
		if err != nil {
			// If Redis is down, allow the request (fail-open)
			return c.Next()
		}

		// Set expiry on the first request in the window
		if count == 1 {
			redisClient.Expire(ctx, key, config.Window)
		}

		// Get remaining TTL for the Retry-After header
		ttl, _ := redisClient.TTL(ctx, key).Result()

		// Set rate limit headers
		c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", config.Max))
		c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, int64(config.Max)-count)))

		if count > int64(config.Max) {
			retryAfter := int(ttl.Seconds())
			if retryAfter <= 0 {
				retryAfter = int(config.Window.Seconds())
			}
			c.Set("Retry-After", fmt.Sprintf("%d", retryAfter))

			return c.Status(fiber.StatusTooManyRequests).JSON(dto.APIResponse{
				Status:     false,
				StatusCode: 429,
				Message:    "Too many requests. Please try again later.",
			})
		}

		return c.Next()
	}
}
