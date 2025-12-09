package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

func AcceptMiddleware(allowedTypes ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		accepted := c.Accepts(allowedTypes...)
		if accepted == "" {
			return c.Status(fiber.StatusNotAcceptable).JSON(fiber.Map{
				"error": "Accept header must be one of: " + strings.Join(allowedTypes, ", "),
			})
		}
		// Optionally store the accepted type in locals
		c.Locals("accepted", accepted)
		return c.Next()
	}
}
