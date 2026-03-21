package controller

import (
	"github.com/231031/wellpaw-backend/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/stripe/stripe-go/v84/webhook"
)

type WebhookController interface {
	HandleAllWebhook(ctx *fiber.Ctx) error
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

func (s *webhookController) HandleAllWebhook(ctx *fiber.Ctx) error {
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

	if err := s.webhookService.MapWebhookHandler(ctxWithTimeOut, &event); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(map[string]interface{}{
			"message": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(map[string]interface{}{
		"message": "success",
	})
}
