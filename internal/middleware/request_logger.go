package middleware

import (
	"time"

	"github.com/ainyx/user-api/internal/logger"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// RequestLogger logs the method, path, status code, duration and request ID
// for every HTTP request using the Zap structured logger.
func RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Process the request
		err := c.Next()

		duration := time.Since(start)
		requestID, _ := c.Locals("requestId").(string)

		logger.Log.Info("HTTP Request",
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", c.Response().StatusCode()),
			zap.Duration("duration", duration),
			zap.String("request_id", requestID),
		)

		return err
	}
}
