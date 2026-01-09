package controller

import (
	"time"

	"github.com/231031/wellpaw-backend/internal/service"
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
// @Description Process image file with OCR
// @tags OCR
// @Accept multipart/form-data
// @Produce application/json
// @Param image formData file true "Image file to process"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /ocr/request [post]
func (c *ocrController) ProcessOcrRequest(ctx *fiber.Ctx) error {
	fileHeader, err := ctx.FormFile("image")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "failed to get image file",
			"error":   err.Error(),
		})
	}

	file, err := fileHeader.Open()
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "failed to open image file",
			"error":   err.Error(),
		})
	}
	defer file.Close()

	ctxWithTimeout, cancel := withTimeout(ctx.Context(), 10*time.Second)
	defer cancel()

	response := c.ocrService.ProcessOcrRequest(ctxWithTimeout, file)
	return ctx.Status(response.Status).JSON(response)
}
