package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/231031/wellpaw-backend/internal/service"
	"github.com/231031/wellpaw-backend/internal/utils"
	"github.com/gofiber/fiber/v2"
)

type DiseaseController interface {
	PredictDisease(ctx *fiber.Ctx) error
	GetPetSkinImagesByUserID(ctx *fiber.Ctx) error
	GetPetSkinImagesByPetID(ctx *fiber.Ctx) error
	LabeledPetSkinDisease(ctx *fiber.Ctx) error
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
// @Success 201 {object} model.PredictPetSkinImageResponse
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

	ctxWithTimeout, cancel := withTimeout(ctx.Context(), 60*time.Second)
	defer cancel()

	response := c.diseaseService.PredictPetSkinDisease(ctxWithTimeout, &payload)
	return ctx.Status(response.Status).JSON(response)
}

// @Summary Get Pet Skin Images
// @Description get all pet skin images of current user
// @tags Disease
// @Security BearerAuth
// @Accept application/json
// @Produce application/json
// @Success 200 {object} model.PetSkinImageResponse
// @Failure 401 {object} model.HTTPResponse
// @Failure 500 {object} model.HTTPResponse
// @Router /diseases [get]
func (c *diseaseController) GetPetSkinImagesByUserID(ctx *fiber.Ctx) error {
	userID := ctx.Locals("id").(uint)

	ctxWithTimeout, cancel := withTimeout(ctx.Context(), defaultTimeout)
	defer cancel()

	response := c.diseaseService.GetPetSkinImagesByUserID(ctxWithTimeout, userID)
	return ctx.Status(response.Status).JSON(response)
}

// @Summary Get Pet Skin Images By Pet ID
// @Description get pet skin images by pet id and current user ownership
// @tags Disease
// @Security BearerAuth
// @Accept application/json
// @Produce application/json
// @Param pet_id path int true "Pet ID"
// @Success 200 {object} model.PetSkinImageResponse
// @Failure 400 {object} model.HTTPResponse
// @Failure 401 {object} model.HTTPResponse
// @Failure 500 {object} model.HTTPResponse
// @Router /diseases/{pet_id} [get]
func (c *diseaseController) GetPetSkinImagesByPetID(ctx *fiber.Ctx) error {
	userID := ctx.Locals("id").(uint)
	rawPetID := ctx.Params("pet_id")
	petID64, err := strconv.ParseUint(rawPetID, 10, 64)
	if err != nil || petID64 == 0 {
		return ctx.Status(http.StatusBadRequest).JSON(model.HTTPResponse{
			Status:  http.StatusBadRequest,
			Message: "invalid pet_id",
		})
	}

	ctxWithTimeout, cancel := withTimeout(ctx.Context(), defaultTimeout)
	defer cancel()

	response := c.diseaseService.GetPetSkinImagesByPetID(ctxWithTimeout, userID, uint(petID64))
	return ctx.Status(response.Status).JSON(response)
}

// @Summary Label Pet Skin Disease
// @Description label latest pet skin image by pet id
// @tags Disease
// @Security BearerAuth
// @Accept application/json
// @Produce application/json
// @Param LabeledPetSkinDiseasePayload body model.LabeledPetSkinDiseasePayload true "Label pet skin disease payload"
// @Success 200 {object} model.PredictPetSkinImageResponse
// @Failure 400 {object} model.HTTPResponse
// @Failure 401 {object} model.HTTPResponse
// @Failure 404 {object} model.HTTPResponse
// @Failure 500 {object} model.HTTPResponse
// @Router /disease/labeled [patch]
func (c *diseaseController) LabeledPetSkinDisease(ctx *fiber.Ctx) error {
	userID := ctx.Locals("id").(uint)

	var payload model.LabeledPetSkinDiseasePayload
	if err := ctx.BodyParser(&payload); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(model.HTTPResponse{
			Status:  http.StatusBadRequest,
			Message: "invalid request body",
		})
	}

	if validationResponse, err := utils.ValidateStruct(&payload); err != nil {
		return ctx.Status(validationResponse.Status).JSON(validationResponse)
	}

	ctxWithTimeout, cancel := withTimeout(ctx.Context(), defaultTimeout)
	defer cancel()

	response := c.diseaseService.LabeledPetSkinDisease(ctxWithTimeout, userID, &payload)
	return ctx.Status(response.Status).JSON(response)
}
