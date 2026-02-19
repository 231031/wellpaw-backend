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
	GetPaymentIntentByID(ctx *fiber.Ctx) error
	StartSubscription(ctx *fiber.Ctx) error
	UpdateSubscription(ctx *fiber.Ctx) error
	CancelSubscription(ctx *fiber.Ctx) error

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

// @Summary Get User All Info
// @Description get user all info
// @tags User
// @Security BearerAuth
// @Accept application/json
// @Produce application/json
// @Success 200 {object} model.UserResponse
// @Failure 400 {object} model.HTTPResponse
// @Failure 500 {object} model.HTTPResponse
// @Router /user [get]
func (c *userController) GetUserAllInfo(ctx *fiber.Ctx) error {
	userID := ctx.Locals("id").(uint)

	ctxWithTimeOut, cancel := withTimeout(ctx.Context(), defaultTimeout)
	defer cancel()

	response := c.userService.GetUserAllInfoByID(ctxWithTimeOut, userID)
	return ctx.Status(response.Status).JSON(response)
}

// @Summary Update Payment Method
// @Description update payment method
// @tags User
// @Security BearerAuth
// @Accept application/json
// @Produce application/json
// @Param   UpdatePaymentMethodPayload body model.PaymentMethodUpdatePayload true "Update payment method payload"
// @Success 200 {object} model.UserResponse
// @Failure 400 {object} model.HTTPResponse
// @Failure 500 {object} model.HTTPResponse
// @Router /user/paymentmethod [patch]
func (c *userController) UpdatePaymentMethod(ctx *fiber.Ctx) error {
	userID := ctx.Locals("id").(uint)

	var payload model.PaymentMethodUpdatePayload
	if err := ctx.BodyParser(&payload); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(model.HTTPResponse{
			Status:  http.StatusBadRequest,
			Message: utils.FailedToUpdateMsg + "payment method",
		})
	}
	if validationResponse, err := utils.ValidateStruct(&payload); err != nil {
		return ctx.Status(validationResponse.Status).JSON(validationResponse)
	}

	ctxWithTimeOut, cancel := withTimeout(ctx.Context(), defaultTimeout)
	defer cancel()

	response := c.userService.UpdatePaymentMethod(ctxWithTimeOut, userID, payload.PaymentMethodID)
	return ctx.Status(response.Status).JSON(response)
}

// @Summary Get All Subscriptions Plan
// @Description get all subscriptions plan and their details provide for customer to select
// @tags User
// @Security BearerAuth
// @Accept application/json
// @Produce application/json
// @Success 200 {object} model.SubscriptionPlanResponse
// @Failure 400 {object} model.HTTPResponse
// @Failure 500 {object} model.HTTPResponse
// @Router /user/subscription [get]
func (c *userController) GetAllSubscriptionsPlan(ctx *fiber.Ctx) error {
	ctxWithTimeOut, cancel := withTimeout(ctx.Context(), defaultTimeout)
	defer cancel()

	response := c.userService.GetAllSubscriptionsPlan(ctxWithTimeOut)
	return ctx.Status(response.Status).JSON(response)
}

// @Summary Get All Subscriptions By Customer ID
// @Description get all subscription histories by customer id
// @tags User
// @Security BearerAuth
// @Accept application/json
// @Produce application/json
// @Param last_id query string false "Last id"
// @Success 200 {object} model.SubscriptionHistoryPaginationResponse
// @Failure 400 {object} model.HTTPResponse
// @Failure 500 {object} model.HTTPResponse
// @Router /user/subscription/history [get]
func (c *userController) GetAllSubscriptionsByCustomerID(ctx *fiber.Ctx) error {
	customerID := ctx.Locals("customer_id").(string)
	lastID := ctx.Query("last_id", "")

	ctxWithTimeOut, cancel := withTimeout(ctx.Context(), 5*time.Second)
	defer cancel()

	response := c.userService.GetAllSubscriptionsByCustomerID(ctxWithTimeOut, customerID, lastID)
	return ctx.Status(response.Status).JSON(response)
}

// @Summary Get Payment Intent By ID
// @Description get payment intent by id to use retry payment in frontend (in case of payment failed)
// @tags User
// @Security BearerAuth
// @Accept application/json
// @Produce application/json
// @Param payment_intent_id path string true "Payment intent id"
// @Success 200 {object} model.PaymentIntentResponse
// @Failure 400 {object} model.HTTPResponse
// @Failure 500 {object} model.HTTPResponse
// @Router /user/paymentintent/{payment_intent_id} [get]
func (c *userController) GetPaymentIntentByID(ctx *fiber.Ctx) error {
	paymentIntentID := ctx.Params("payment_intent_id")

	ctxWithTimeOut, cancel := withTimeout(ctx.Context(), defaultTimeout)
	defer cancel()

	if paymentIntentID == "" {
		return ctx.Status(http.StatusBadRequest).JSON(model.HTTPResponse{
			Status:  http.StatusBadRequest,
			Message: "payment intent id is required",
		})
	}
	response := c.userService.GetPaymentIntentByID(ctxWithTimeOut, paymentIntentID)
	return ctx.Status(response.Status).JSON(response)
}

// @Summary Start Subscription
// @Description start subscription after attached payment method and return payment intent detail to confirm payment in frontend
// @tags User
// @Security BearerAuth
// @Accept application/json
// @Produce application/json
// @Param   StartSubscriptionPayload body model.StartSubscriptionPayload true "Start subscription payload"
// @Success 200 {object} model.PaymentIntentResponse
// @Failure 400 {object} model.HTTPResponse
// @Failure 500 {object} model.HTTPResponse
// @Router /user/subscription/start [post]
func (c *userController) StartSubscription(ctx *fiber.Ctx) error {
	userID := ctx.Locals("id").(uint)

	var payload model.StartSubscriptionPayload
	if err := ctx.BodyParser(&payload); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(model.HTTPResponse{
			Status:  http.StatusBadRequest,
			Message: utils.FailedToUpdateMsg + "payment method",
		})
	}
	if validationResponse, err := utils.ValidateStruct(&payload); err != nil {
		return ctx.Status(validationResponse.Status).JSON(validationResponse)
	}

	ctxWithTimeOut, cancel := withTimeout(ctx.Context(), 10*time.Second)
	defer cancel()

	response := c.userService.StartSubscription(ctxWithTimeOut, userID, payload)
	return ctx.Status(response.Status).JSON(response)
}

// @Summary Update Subscription
// @Description update subscription return payment intent detail to confirm payment in frontend
// @tags User
// @Security BearerAuth
// @Accept application/json
// @Produce application/json
// @Param   UpdateSubscriptionPayload body model.UpdateSubscriptionPayload true "Update subscription payload"
// @Success 200 {object} model.PaymentIntentResponse
// @Failure 400 {object} model.HTTPResponse
// @Failure 500 {object} model.HTTPResponse
// @Router /user/subscription/update [patch]
func (c *userController) UpdateSubscription(ctx *fiber.Ctx) error {
	customerID := ctx.Locals("customer_id").(string)

	var payload model.UpdateSubscriptionPayload
	if err := ctx.BodyParser(&payload); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(model.HTTPResponse{
			Status:  http.StatusBadRequest,
			Message: utils.FailedToUpdateMsg + "subscription",
		})
	}
	if validationResponse, err := utils.ValidateStruct(&payload); err != nil {
		return ctx.Status(validationResponse.Status).JSON(validationResponse)
	}

	ctxWithTimeOut, cancel := withTimeout(ctx.Context(), 5*time.Second)
	defer cancel()

	response := c.userService.UpdateSubscription(ctxWithTimeOut, customerID, payload)
	return ctx.Status(response.Status).JSON(response)
}

// @Summary Cancel Subscription
// @Description cancel subscription
// @tags User
// @Security BearerAuth
// @Accept application/json
// @Produce application/json
// @Success 200 {object} model.SubscriptionHistoryResponse
// @Failure 400 {object} model.HTTPResponse
// @Failure 500 {object} model.HTTPResponse
// @Router /user/subscription/cancel/{subscription_id} [get]
func (c *userController) CancelSubscription(ctx *fiber.Ctx) error {
	subscriptionID := ctx.Params("subscription_id")

	ctxWithTimeOut, cancel := withTimeout(ctx.Context(), 5*time.Second)
	defer cancel()

	response := c.userService.CancelSubscription(ctxWithTimeOut, subscriptionID)
	return ctx.Status(response.Status).JSON(response)
}

// @Summary Manage Food Notification
// @Description manage food notification
// @tags User
// @Security BearerAuth
// @Accept application/json
// @Produce application/json
// @Success 200 {object} model.UserResponse
// @Failure 400 {object} model.HTTPResponse
// @Failure 500 {object} model.HTTPResponse
// @Router /user/notification/food [get]
func (c *userController) ManageFoodNotification(ctx *fiber.Ctx) error {
	userID := ctx.Locals("id").(uint)

	ctxWithTimeOut, cancel := withTimeout(ctx.Context(), defaultTimeout)
	defer cancel()

	response := c.userService.ManageFoodNotification(ctxWithTimeOut, userID)
	return ctx.Status(response.Status).JSON(response)
}

// @Summary Manage Calendar Notification
// @Description manage calendar notification
// @tags User
// @Security BearerAuth
// @Accept application/json
// @Produce application/json
// @Success 200 {object} model.UserResponse
// @Failure 400 {object} model.HTTPResponse
// @Failure 500 {object} model.HTTPResponse
// @Router /user/notification/calendar [get]
func (c *userController) ManageCalendarNotification(ctx *fiber.Ctx) error {
	userID := ctx.Locals("id").(uint)

	ctxWithTimeOut, cancel := withTimeout(ctx.Context(), defaultTimeout)
	defer cancel()

	response := c.userService.ManageCalendarNotification(ctxWithTimeOut, userID)
	return ctx.Status(response.Status).JSON(response)
}
