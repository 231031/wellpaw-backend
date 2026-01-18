package controller

import (
	"net/http"
	"time"

	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/231031/wellpaw-backend/internal/service"
	"github.com/231031/wellpaw-backend/internal/utils"
	"github.com/gofiber/fiber/v2"
)

type UserController interface {
	GetUserAllInfo(ctx *fiber.Ctx) error
	UpdatePaymentMethod(ctx *fiber.Ctx) error

	GetAllSubscriptionsPlan(ctx *fiber.Ctx) error

	GetAllSubscriptionsByCustomerID(ctx *fiber.Ctx) error
	StartSubscription(ctx *fiber.Ctx) error
	UpdateSubscription(ctx *fiber.Ctx) error

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

func (c *userController) GetAllSubscriptionsPlan(ctx *fiber.Ctx) error {
	ctxWithTimeOut, cancel := withTimeout(ctx.Context(), defaultTimeout)
	defer cancel()

	response := c.userService.GetAllSubscriptionsPlan(ctxWithTimeOut)
	return ctx.Status(response.Status).JSON(response)
}

func (c *userController) GetAllSubscriptionsByCustomerID(ctx *fiber.Ctx) error {
	customerID := ctx.Locals("customer_id").(string)

	ctxWithTimeOut, cancel := withTimeout(ctx.Context(), defaultTimeout)
	defer cancel()

	response := c.userService.GetAllSubscriptionsByCustomerID(ctxWithTimeOut, customerID)
	return ctx.Status(response.Status).JSON(response)
}

func (c *userController) StartSubscription(ctx *fiber.Ctx) error {
	userID := ctx.Locals("id").(uint)

	var payload model.StartSubscriptionPayload
	if err := ctx.BodyParser(&payload); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(model.HTTPResponse{
			Status:  http.StatusBadRequest,
			Message: utils.FailedToUpdateMsg + "payment method",
		})
	}

	ctxWithTimeOut, cancel := withTimeout(ctx.Context(), 10*time.Second)
	defer cancel()

	response := c.userService.StartSubscription(ctxWithTimeOut, userID, payload)
	return ctx.Status(response.Status).JSON(response)
}

func (c *userController) UpdateSubscription(ctx *fiber.Ctx) error {
	customerID := ctx.Locals("customer_id").(string)

	var payload model.UpdateSubscriptionPayload
	if err := ctx.BodyParser(&payload); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(model.HTTPResponse{
			Status:  http.StatusBadRequest,
			Message: utils.FailedToUpdateMsg + "subscription",
		})
	}

	ctxWithTimeOut, cancel := withTimeout(ctx.Context(), 5*time.Second)
	defer cancel()

	response := c.userService.UpdateSubscription(ctxWithTimeOut, customerID, payload)
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
