package controller

import (
	"time"

	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/231031/wellpaw-backend/internal/service"
	"github.com/231031/wellpaw-backend/internal/utils"
	"github.com/gofiber/fiber/v2"
)

type OcrController interface {
	ProcessOcrRequest(ctx *fiber.Ctx) error
}

type ocrController struct {
	ocrService service.OcrService
}

func NewOcrController(ocrService service.OcrService) OcrController {
	return &ocrController{
		ocrService: ocrService,
	}
}

// @Summary Request OCR
// @Description Process base64 image with OCR
// @tags OCR
// @Security BearerAuth
// @Accept application/json
// @Produce application/json
// @Param OcrRequestPayload body model.OcrRequestPayload true "OCR request payload"
// @Success 200 {object} model.OcrPetFoodResponse
// @Failure 400 {object} model.HTTPResponse
// @Failure 500 {object} model.HTTPResponse
// @Router /ocr/request [post]
func (c *ocrController) ProcessOcrRequest(ctx *fiber.Ctx) error {
	userID := ctx.Locals("id").(uint)

	var payload model.OcrRequestPayload
	if err := ctx.BodyParser(&payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(model.HTTPResponse{
			Status:  fiber.StatusBadRequest,
			Message: "invalid request body",
		})
	}

	if validationResponse, err := utils.ValidateStruct(&payload); err != nil {
		return ctx.Status(validationResponse.Status).JSON(validationResponse)
	}

	ctxWithTimeout, cancel := withTimeout(ctx.Context(), 60*time.Second)
	defer cancel()

	response := c.ocrService.ProcessOcrRequest(ctxWithTimeout, userID, payload.Image)
	return ctx.Status(response.Status).JSON(response)
}
