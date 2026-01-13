package controller

import (
	"net/http"

	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/231031/wellpaw-backend/internal/service"
	"github.com/231031/wellpaw-backend/internal/utils"
	"github.com/gofiber/fiber/v2"
)

type UserController interface {
	GetUserAllInfo(ctx *fiber.Ctx) error
	UpdatePaymentMethod(ctx *fiber.Ctx) error
	ManageFoodNotification(ctx *fiber.Ctx) error
	ManageCalendarNotification(ctx *fiber.Ctx) error
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

func (c *userController) UpdatePaymentMethod(ctx *fiber.Ctx) error {
	userID := ctx.Locals("id").(uint)

	var payload model.PaymentMethodUpdatePayload
	if err := ctx.BodyParser(&payload); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(model.HTTPResponse{
			Status:  http.StatusBadRequest,
			Message: utils.FailedToUpdateMsg + "payment method",
		})
	}

	ctxWithTimeOut, cancel := withTimeout(ctx.Context(), defaultTimeout)
	defer cancel()

	response := c.userService.UpdatePaymentMethod(ctxWithTimeOut, userID, payload.PaymentMethodID)
	return ctx.Status(response.Status).JSON(response)
}

func (c *userController) ManageFoodNotification(ctx *fiber.Ctx) error {
	userID := ctx.Locals("id").(uint)

	ctxWithTimeOut, cancel := withTimeout(ctx.Context(), defaultTimeout)
	defer cancel()

	response := c.userService.ManageFoodNotification(ctxWithTimeOut, userID)
	return ctx.Status(response.Status).JSON(response)
}

func (c *userController) ManageCalendarNotification(ctx *fiber.Ctx) error {
	userID := ctx.Locals("id").(uint)

	ctxWithTimeOut, cancel := withTimeout(ctx.Context(), defaultTimeout)
	defer cancel()

	response := c.userService.ManageCalendarNotification(ctxWithTimeOut, userID)
	return ctx.Status(response.Status).JSON(response)
}
