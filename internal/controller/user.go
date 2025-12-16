package controller

import (
	"github.com/231031/pethealth-backend/internal/service"
	"github.com/gofiber/fiber/v2"
)

type UserController interface {
	GetUserAllInfo(ctx *fiber.Ctx) error
}

type userController struct {
	userService service.UserService
}

func NewUserController(userService service.UserService) UserController {
	return &userController{
		userService: userService,
	}
}

func (c *userController) GetUserAllInfo(ctx *fiber.Ctx) error {
	userID := ctx.Locals("id").(uint)

	ctxWithTimeOut, cancel := withTimeout(ctx.Context(), defaultTimeout)
	defer cancel()

	response := c.userService.GetUserAllInfoByID(ctxWithTimeOut, userID)
	return ctx.Status(response.Status).JSON(response)
}
