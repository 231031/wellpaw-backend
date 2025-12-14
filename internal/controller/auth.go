package controller

import (
	"github.com/231031/pethealth-backend/internal/model"
	"github.com/231031/pethealth-backend/internal/service"
	"github.com/gofiber/fiber/v2"
)

type AuthController interface {
	CreateUser(ctx *fiber.Ctx) error
	LoginUser(ctx *fiber.Ctx) error
}

type authController struct {
	authService service.AuthService
}

func NewAuthController(authService service.AuthService) AuthController {
	return &authController{
		authService: authService,
	}
}

func (c *authController) CreateUser(ctx *fiber.Ctx) error {
	var user model.User
	if err := ctx.BodyParser(&user); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "invalid request body",
		})
	}

	// validate fields

	response := c.authService.CreateUser(ctx.Context(), &user)
	return ctx.Status(response.Status).JSON(response)
}

func (c *authController) LoginUser(ctx *fiber.Ctx) error {
	var payload model.LoginPayload
	if err := ctx.BodyParser(&payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "invalid request body",
		})
	}

	// validate fields

	response := c.authService.LoginUser(ctx.Context(), &payload)
	return ctx.Status(response.Status).JSON(response)
}
