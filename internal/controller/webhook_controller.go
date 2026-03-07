package controller

import (
	"github.com/231031/wellpaw-backend/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/stripe/stripe-go/v84/webhook"
)

type WebhookController interface {
	HandleSubscriptionUpdated(ctx *fiber.Ctx) error
	HandleSubscriptionFreeTiralUpdated(ctx *fiber.Ctx) error
}

type webhookController struct {
	webhookSecretKey string
	webhookService   service.WebhookService
}

func NewWebhookController(webhookSecretKey string, webhookService service.WebhookService) WebhookController {
	return &webhookController{
		webhookSecretKey: webhookSecretKey,
		webhookService:   webhookService,
	}
}

func (s *webhookController) HandleSubscriptionUpdated(ctx *fiber.Ctx) error {
	ctxWithTimeOut, cancel := withTimeout(ctx.Context(), defaultTimeout)
	defer cancel()

	payload := ctx.Request().Body()
	signatureHeader := ctx.Get("Stripe-Signature")
	event, err := webhook.ConstructEvent(payload, signatureHeader, s.webhookSecretKey)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(map[string]interface{}{
			"message": err.Error(),
		})
	}

	if err := s.webhookService.HandleSubscriptionUpdated(ctxWithTimeOut, &event); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(map[string]interface{}{
			"message": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(map[string]interface{}{
		"message": "success",
	})
}

func (s *webhookController) HandleSubscriptionFreeTiralUpdated(ctx *fiber.Ctx) error {
	ctxWithTimeOut, cancel := withTimeout(ctx.Context(), defaultTimeout)
	defer cancel()

	payload := ctx.Request().Body()
	signatureHeader := ctx.Get("Stripe-Signature")
	event, err := webhook.ConstructEvent(payload, signatureHeader, s.webhookSecretKey)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(map[string]interface{}{
			"message": err.Error(),
		})
	}

	if err := s.webhookService.HandleSubscriptionFreeTierUpdated(ctxWithTimeOut, &event); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(map[string]interface{}{
			"message": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(map[string]interface{}{
		"message": "success",
	})
}
