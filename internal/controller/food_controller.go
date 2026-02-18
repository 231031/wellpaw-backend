package controller

import (
	"net/http"

	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/231031/wellpaw-backend/internal/service"
	"github.com/gofiber/fiber/v2"
)

type FoodController interface {
	CreateFood(ctx *fiber.Ctx) error
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

	if payload.Type < model.DRY || payload.Type > model.SUPPLEMENTS {
		return ctx.Status(http.StatusBadRequest).JSON(model.HTTPResponse{
			Status:  http.StatusBadRequest,
			Message: "invalid food type",
		})
	}

	payload.UserID = userID
	ctxWithTimeout, cancel := withTimeout(ctx.Context(), defaultTimeout)
	defer cancel()

	response := c.foodService.CreateFood(ctxWithTimeout, &payload)
	return ctx.Status(response.Status).JSON(response)
}
