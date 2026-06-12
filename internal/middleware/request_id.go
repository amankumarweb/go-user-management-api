package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// RequestID generates a unique X-Request-ID for every incoming request.
// If the client already provides the header, it is preserved.
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Set the header on the response so the client can correlate.
		c.Set("X-Request-ID", requestID)

		// Store in Locals so downstream middleware (e.g. request logger) can read it.
		c.Locals("requestId", requestID)

		return c.Next()
	}
}
