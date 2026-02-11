package controller

import (
	"net/http"

	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/231031/wellpaw-backend/internal/service"
	"github.com/gofiber/fiber/v2"
)

type PetController interface {
	CreateNewPet(ctx *fiber.Ctx) error
	UpdatePetInfo(ctx *fiber.Ctx) error
	UpdatePetDetail(ctx *fiber.Ctx) error
}

type petController struct {
	petService service.PetService
}

func NewPetController(petService service.PetService) PetController {
	return &petController{
		petService: petService,
	}
}

// @Summary Create New Pet
// @Description create new pet with pet info and pet detail
// @tags Pet
// @Security BearerAuth
// @Accept application/json
// @Produce application/json
// @Param   PetPayload body model.PetPayload true "Create pet payload"
// @Success 201 {object} model.PetResponse
// @Failure 400 {object} model.HTTPResponse
// @Failure 401 {object} model.HTTPResponse
// @Failure 500 {object} model.HTTPResponse
// @Router /pet [post]
func (c *petController) CreateNewPet(ctx *fiber.Ctx) error {
	userID := ctx.Locals("id").(uint)

	var payload model.PetPayload
	if err := ctx.BodyParser(&payload); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(model.HTTPResponse{
			Status:  http.StatusBadRequest,
			Message: "invalid request body",
		})
	}

	payload.PetInfo.UserID = userID
	ctxWithTimeout, cancel := withTimeout(ctx.Context(), defaultTimeout)
	defer cancel()

	response := c.petService.CreateNewPet(ctxWithTimeout, &payload)
	return ctx.Status(response.Status).JSON(response)
}

// @Summary Update Pet Info
// @Description update pet base info
// @tags Pet
// @Security BearerAuth
// @Accept application/json
// @Produce application/json
// @Param   UpdatePetInfoPayload body model.Pet true "Update pet info payload"
// @Success 200 {object} model.PetResponse
// @Failure 400 {object} model.HTTPResponse
// @Failure 401 {object} model.HTTPResponse
// @Failure 404 {object} model.HTTPResponse
// @Failure 500 {object} model.HTTPResponse
// @Router /pet/info [patch]
func (c *petController) UpdatePetInfo(ctx *fiber.Ctx) error {
	userID := ctx.Locals("id").(uint)

	var payload model.Pet
	if err := ctx.BodyParser(&payload); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(model.HTTPResponse{
			Status:  http.StatusBadRequest,
			Message: "invalid request body",
		})
	}

	payload.UserID = userID
	ctxWithTimeout, cancel := withTimeout(ctx.Context(), defaultTimeout)
	defer cancel()

	response := c.petService.UpdatePetInfo(ctxWithTimeout, &payload)
	return ctx.Status(response.Status).JSON(response)
}

// @Summary Update Pet Detail
// @Description update pet detail and recalculate nutrients
// @tags Pet
// @Security BearerAuth
// @Accept application/json
// @Produce application/json
// @Param   UpdatePetDetailPayload body model.PetDetail true "Update pet detail payload"
// @Success 200 {object} model.PetResponse
// @Failure 400 {object} model.HTTPResponse
// @Failure 401 {object} model.HTTPResponse
// @Failure 404 {object} model.HTTPResponse
// @Failure 500 {object} model.HTTPResponse
// @Router /pet/detail [post]
func (c *petController) UpdatePetDetail(ctx *fiber.Ctx) error {
	var payload model.PetDetail
	if err := ctx.BodyParser(&payload); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(model.HTTPResponse{
			Status:  http.StatusBadRequest,
			Message: "invalid request body",
		})
	}

	ctxWithTimeout, cancel := withTimeout(ctx.Context(), defaultTimeout)
	defer cancel()

	response := c.petService.UpdatePetDetail(ctxWithTimeout, &payload)
	return ctx.Status(response.Status).JSON(response)
}
