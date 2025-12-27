package middleware

import (
	"strings"

	"github.com/231031/wellpaw-backend/internal/service"
	"github.com/231031/wellpaw-backend/internal/utils"
	"github.com/gofiber/fiber/v2"
)

type AuthMiddleware interface {
	AuthorizeUser() fiber.Handler
}

type authMiddleware struct {
	tokenService service.TokenService
}

func NewAuthMiddleware(tokenService service.TokenService) AuthMiddleware {
	return &authMiddleware{
		tokenService: tokenService,
	}
}

func (m *authMiddleware) AuthorizeUser() fiber.Handler {
	return func(c *fiber.Ctx) error {
		headers := c.GetReqHeaders()
		var auth []string
		for key, val := range headers {
			if key == "Authorization" {
				auth = val
			}
		}
		if auth == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": utils.ErrUnauthHeader.Error(),
			})
		}

		tokenStr := strings.Split(auth[0], "Bearer ")
		if tokenStr == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": utils.ErrUnauthHeader.Error(),
			})
		}
		if len(tokenStr) < 2 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": utils.ErrUnauthHeader.Error(),
			})
		}

		// validate token and extract the id and role and check with the data on redis
		claims, err := m.tokenService.ValidateToken(tokenStr[1])
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": utils.ErrUnauthHeader.Error(),
			})
		}

		c.Locals("id", claims.User.ID)
		c.Locals("tier", claims.User.Tier)

		return c.Next()
	}
}
