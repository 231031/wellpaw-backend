package controller

import (
	"net/http"
	"strconv"

	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/231031/wellpaw-backend/internal/service"
	"github.com/231031/wellpaw-backend/internal/utils"
	"github.com/gofiber/fiber/v2"
)

type FoodController interface {
	CreateFood(ctx *fiber.Ctx) error
	GetFoodsByUserID(ctx *fiber.Ctx) error
	GetFoodsByFoodType(ctx *fiber.Ctx) error
	UpdateFoodDetail(ctx *fiber.Ctx) error
	SoftDeleteFood(ctx *fiber.Ctx) error
}

type foodController struct {
	foodService service.FoodService
}

func NewFoodController(foodService service.FoodService) FoodController {
	return &foodController{
		foodService: foodService,
	}
}

// @Summary Create Food
// @Description create food for the authorized user
// @tags Food
// @Security BearerAuth
// @Accept application/json
// @Produce application/json
// @Param   CreateFoodPayload body model.Food true "Create food payload"
// @Success 201 {object} model.FoodResponse
// @Failure 400 {object} model.HTTPResponse
// @Failure 401 {object} model.HTTPResponse
// @Failure 500 {object} model.HTTPResponse
// @Router /food [post]
func (c *foodController) CreateFood(ctx *fiber.Ctx) error {
	userID := ctx.Locals("id").(uint)

	var payload model.Food
	if err := ctx.BodyParser(&payload); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(model.HTTPResponse{
			Status:  http.StatusBadRequest,
			Message: "invalid request body",
		})
	}

	payload.UserID = userID
	if validationResponse, err := utils.ValidateStruct(&payload); err != nil {
		return ctx.Status(validationResponse.Status).JSON(validationResponse)
	}

	ctxWithTimeout, cancel := withTimeout(ctx.Context(), defaultTimeout)
	defer cancel()

	response := c.foodService.CreateFood(ctxWithTimeout, &payload)
	return ctx.Status(response.Status).JSON(response)
}

// @Summary Get Foods
// @Description get all foods of current user
// @tags Food
// @Security BearerAuth
// @Accept application/json
// @Produce application/json
// @Success 200 {object} model.FoodsResponse
// @Failure 401 {object} model.HTTPResponse
// @Failure 500 {object} model.HTTPResponse
// @Router /foods [get]
func (c *foodController) GetFoodsByUserID(ctx *fiber.Ctx) error {
	userID := ctx.Locals("id").(uint)

	ctxWithTimeout, cancel := withTimeout(ctx.Context(), defaultTimeout)
	defer cancel()

	response := c.foodService.GetFoodsByUserID(ctxWithTimeout, userID)
	return ctx.Status(response.Status).JSON(response)
}

// @Summary Get Foods By Type
// @Description get foods of current user by food type (0=dry,1=wet,2=treats,3=supplements)
// @tags Food
// @Security BearerAuth
// @Accept application/json
// @Produce application/json
// @Param food_type path int true "Food type"
// @Success 200 {object} model.FoodsResponse
// @Failure 400 {object} model.HTTPResponse
// @Failure 401 {object} model.HTTPResponse
// @Failure 500 {object} model.HTTPResponse
// @Router /foods/{food_type} [get]
func (c *foodController) GetFoodsByFoodType(ctx *fiber.Ctx) error {
	userID := ctx.Locals("id").(uint)
	rawFoodType := ctx.Params("food_type")
	foodType, err := strconv.Atoi(rawFoodType)
	if err != nil || foodType < int(model.DRY) || foodType > int(model.SUPPLEMENTS) {
		return ctx.Status(http.StatusBadRequest).JSON(model.HTTPResponse{
			Status:  http.StatusBadRequest,
			Message: "invalid food_type",
		})
	}

	ctxWithTimeout, cancel := withTimeout(ctx.Context(), defaultTimeout)
	defer cancel()

	response := c.foodService.GetFoodsByFoodType(ctxWithTimeout, userID, model.FoodType(foodType))
	return ctx.Status(response.Status).JSON(response)
}

// @Summary Update Food Detail
// @Description update food fields (weight, quantity/quality, name, image_path)
// @tags Food
// @Security BearerAuth
// @Accept application/json
// @Produce application/json
// @Param   UpdateFoodDetailPayload body model.UpdateFoodDetailPayload true "Update food payload"
// @Success 200 {object} model.FoodResponse
// @Failure 400 {object} model.HTTPResponse
// @Failure 401 {object} model.HTTPResponse
// @Failure 404 {object} model.HTTPResponse
// @Failure 500 {object} model.HTTPResponse
// @Router /food [patch]
func (c *foodController) UpdateFoodDetail(ctx *fiber.Ctx) error {
	userID := ctx.Locals("id").(uint)

	var payload model.UpdateFoodDetailPayload
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

	response := c.foodService.UpdateFoodDetail(ctxWithTimeout, userID, payload.FoodID, &payload)
	return ctx.Status(response.Status).JSON(response)
}

// @Summary Soft Delete Food
// @Description soft delete food by id
// @tags Food
// @Security BearerAuth
// @Accept application/json
// @Produce application/json
// @Param food_id path int true "Food ID"
// @Success 200 {object} model.HTTPResponse
// @Failure 400 {object} model.HTTPResponse
// @Failure 401 {object} model.HTTPResponse
// @Failure 404 {object} model.HTTPResponse
// @Failure 500 {object} model.HTTPResponse
// @Router /food/{food_id} [delete]
func (c *foodController) SoftDeleteFood(ctx *fiber.Ctx) error {
	userID := ctx.Locals("id").(uint)
	rawFoodID := ctx.Params("food_id")
	foodID64, err := strconv.ParseUint(rawFoodID, 10, 64)
	if err != nil || foodID64 == 0 {
		return ctx.Status(http.StatusBadRequest).JSON(model.HTTPResponse{
			Status:  http.StatusBadRequest,
			Message: "invalid food_id",
		})
	}

	ctxWithTimeout, cancel := withTimeout(ctx.Context(), defaultTimeout)
	defer cancel()

	response := c.foodService.SoftDeleteFood(ctxWithTimeout, userID, uint(foodID64))
	return ctx.Status(response.Status).JSON(response)
}
