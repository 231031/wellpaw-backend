package controller

import (
	"net/http"

	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/231031/wellpaw-backend/internal/service"
	"github.com/231031/wellpaw-backend/internal/utils"
	"github.com/gofiber/fiber/v2"
)

type DiseaseController interface {
	PredictDisease(ctx *fiber.Ctx) error
}

type diseaseController struct {
	diseaseService service.DiseaseService
}

func NewDiseaseController(diseaseService service.DiseaseService) DiseaseController {
	return &diseaseController{
		diseaseService: diseaseService,
	}
}

// @Summary Predict Disease
// @Description predict pet skin disease from base64 image
// @tags Disease
// @Security BearerAuth
// @Accept application/json
// @Produce application/json
// @Param PredictPetSkinDiseasePayload body model.PredictPetSkinDiseasePayload true "Predict disease payload"
// @Success 201 {object} model.PetSkinImageResponse
// @Failure 400 {object} model.HTTPResponse
// @Failure 401 {object} model.HTTPResponse
// @Failure 404 {object} model.HTTPResponse
// @Failure 500 {object} model.HTTPResponse
// @Router /disease/predict [post]
func (c *diseaseController) PredictDisease(ctx *fiber.Ctx) error {
	var payload model.PredictPetSkinDiseasePayload
	if err := ctx.BodyParser(&payload); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(model.HTTPResponse{
			Status:  http.StatusBadRequest,
			Message: "invalid request body",
		})
	}

	if validationResponse, err := utils.ValidateStruct(&payload); err != nil {
		return ctx.Status(validationResponse.Status).JSON(validationResponse)
	}

	ctxWithTimeout, cancel := withTimeout(ctx.Context(), 60)
	defer cancel()

	response := c.diseaseService.PredictPetSkinDisease(ctxWithTimeout, &payload)
	return ctx.Status(response.Status).JSON(response)
}
