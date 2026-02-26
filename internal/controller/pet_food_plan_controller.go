package controller

import (
	"net/http"
	"strconv"

	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/231031/wellpaw-backend/internal/service"
	"github.com/231031/wellpaw-backend/internal/utils"
	"github.com/gofiber/fiber/v2"
)

type PetFoodPlanController interface {
	CalculatePetFoodPlan(ctx *fiber.Ctx) error
	CreatePetFoodPlan(ctx *fiber.Ctx) error
	GetLastestActivePlanDetailByPet(ctx *fiber.Ctx) error
	UpdateFeedingAmountFromUser(ctx *fiber.Ctx) error
}

type petFoodPlanController struct {
	petFoodPlanService service.PetFoodPlanService
}

func NewPetFoodPlanController(petFoodPlanService service.PetFoodPlanService) PetFoodPlanController {
	return &petFoodPlanController{
		petFoodPlanService: petFoodPlanService,
	}
}

// @Summary Calculate Pet Food Plan
// @Description calculate pet food plan with selected foods and optional grams per cup without persisting
// @tags Pet Food Plan
// @Security BearerAuth
// @Accept application/json
// @Produce application/json
// @Param   CalculatePetFoodPlanPayload body model.CalculatePetFoodPlanPayload true "Calculate pet food plan payload"
// @Success 200 {object} model.PetFoodPlanResponse
// @Failure 400 {object} model.HTTPResponse
// @Failure 401 {object} model.HTTPResponse
// @Failure 404 {object} model.HTTPResponse
// @Failure 500 {object} model.HTTPResponse
// @Router /foodplan/calculate [post]
func (c *petFoodPlanController) CalculatePetFoodPlan(ctx *fiber.Ctx) error {
	userID := ctx.Locals("id").(uint)

	var payload model.CalculatePetFoodPlanPayload
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

	response := c.petFoodPlanService.CalculatePetFoodPlan(ctxWithTimeout, userID, &payload)
	return ctx.Status(response.Status).JSON(response)
}

// @Summary Create Pet Food Plan
// @Description create pet food plan with selected foods and optional grams per cup
// @tags Pet Food Plan
// @Security BearerAuth
// @Accept application/json
// @Produce application/json
// @Param   CreatePetFoodPlanPayload body model.CreatePetFoodPlanPayload true "Create pet food plan payload"
// @Success 201 {object} model.PetFoodPlanResponse
// @Failure 400 {object} model.HTTPResponse
// @Failure 401 {object} model.HTTPResponse
// @Failure 404 {object} model.HTTPResponse
// @Failure 500 {object} model.HTTPResponse
// @Router /foodplan [post]
func (c *petFoodPlanController) CreatePetFoodPlan(ctx *fiber.Ctx) error {
	userID := ctx.Locals("id").(uint)

	var payload model.CreatePetFoodPlanPayload
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

	response := c.petFoodPlanService.CreatePetFoodPlan(ctxWithTimeout, userID, &payload)
	return ctx.Status(response.Status).JSON(response)
}

// @Summary Get Latest Active Pet Food Plan Detail
// @Description get latest active pet food plan detail by pet id
// @tags Pet Food Plan
// @Security BearerAuth
// @Accept application/json
// @Produce application/json
// @Param   pet_id path int true "Pet ID"
// @Success 200 {object} model.PetFoodPlanResponse
// @Failure 400 {object} model.HTTPResponse
// @Failure 401 {object} model.HTTPResponse
// @Failure 404 {object} model.HTTPResponse
// @Failure 500 {object} model.HTTPResponse
// @Router /foodplan/{pet_id} [get]
func (c *petFoodPlanController) GetLastestActivePlanDetailByPet(ctx *fiber.Ctx) error {
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

	response := c.petFoodPlanService.GetLastestActivePlanDetailByPet(ctxWithTimeout, uint(petID64))
	return ctx.Status(response.Status).JSON(response)
}

// @Summary Update Feeding Amount In Pet Food Plan
// @Description update feeding amount in latest active pet food plan and recalculate total intake
// @tags Pet Food Plan
// @Security BearerAuth
// @Accept application/json
// @Produce application/json
// @Param   AdjustAmountFoodInPetFoodPlanPayload body model.AdjustAmountFoodInPetFoodPlanPayload true "Adjust feeding amount payload"
// @Success 200 {object} model.PetFoodPlanResponse
// @Failure 400 {object} model.HTTPResponse
// @Failure 401 {object} model.HTTPResponse
// @Failure 404 {object} model.HTTPResponse
// @Failure 500 {object} model.HTTPResponse
// @Router /foodplan/amount [put]
func (c *petFoodPlanController) UpdateFeedingAmountFromUser(ctx *fiber.Ctx) error {
	var payload model.AdjustAmountFoodInPetFoodPlanPayload
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

	response := c.petFoodPlanService.UpdateFeedingAmountFromUser(ctxWithTimeout, &payload)
	return ctx.Status(response.Status).JSON(response)
}
