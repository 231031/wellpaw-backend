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
