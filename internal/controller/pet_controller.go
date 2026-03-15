package controller

import (
	"net/http"
	"strconv"

	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/231031/wellpaw-backend/internal/service"
	"github.com/231031/wellpaw-backend/internal/utils"
	"github.com/gofiber/fiber/v2"
)

type PetController interface {
	CreateNewPet(ctx *fiber.Ctx) error
	GetPetByPetID(ctx *fiber.Ctx) error
	GetPetsByUserID(ctx *fiber.Ctx) error
	GetPetAnalysisByPetID(ctx *fiber.Ctx) error
	UpdatePetInfo(ctx *fiber.Ctx) error
	UpdatePetDetail(ctx *fiber.Ctx) error
	SoftDeletePet(ctx *fiber.Ctx) error
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
	if validationResponse, err := utils.ValidateStruct(&payload); err != nil {
		return ctx.Status(validationResponse.Status).JSON(validationResponse)
	}
	ctxWithTimeout, cancel := withTimeout(ctx.Context(), defaultTimeout)
	defer cancel()

	response := c.petService.CreateNewPet(ctxWithTimeout, &payload)
	return ctx.Status(response.Status).JSON(response)
}

// @Summary Get User Pet By Pet ID
// @Description get pet by id
// @tags Pet
// @Security BearerAuth
// @Accept application/json
// @Produce application/json
// @Success 200 {object} model.PetResponse
// @Failure 401 {object} model.HTTPResponse
// @Failure 500 {object} model.HTTPResponse
// @Router /pet/{pet_id} [get]
func (c *petController) GetPetByPetID(ctx *fiber.Ctx) error {
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

	response := c.petService.GetPetByID(ctxWithTimeout, uint(petID64))
	return ctx.Status(response.Status).JSON(response)
}

// @Summary Get User Pets
// @Description get all pets by current user
// @tags Pet
// @Security BearerAuth
// @Accept application/json
// @Produce application/json
// @Success 200 {object} model.PetsResponse
// @Failure 401 {object} model.HTTPResponse
// @Failure 500 {object} model.HTTPResponse
// @Router /pets [get]
func (c *petController) GetPetsByUserID(ctx *fiber.Ctx) error {
	userID := ctx.Locals("id").(uint)

	ctxWithTimeout, cancel := withTimeout(ctx.Context(), defaultTimeout)
	defer cancel()

	response := c.petService.GetPetsByUserID(ctxWithTimeout, userID)
	return ctx.Status(response.Status).JSON(response)
}

// @Summary Get Pet Analysis
// @Description get pet monthly analysis by pet id
// @tags Pet
// @Security BearerAuth
// @Accept application/json
// @Produce application/json
// @Param pet_id path int true "Pet ID"
// @Success 200 {object} model.PetPlanAnalysisResponse
// @Failure 400 {object} model.HTTPResponse
// @Failure 401 {object} model.HTTPResponse
// @Failure 404 {object} model.HTTPResponse
// @Failure 500 {object} model.HTTPResponse
// @Router /pet/analysis/{pet_id} [get]
func (c *petController) GetPetAnalysisByPetID(ctx *fiber.Ctx) error {
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

	response := c.petService.GetPetAnalysisByPetID(ctxWithTimeout, uint(petID64))
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
// @Router /pet/info [put]
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
	if validationResponse, err := utils.ValidateStruct(&payload); err != nil {
		return ctx.Status(validationResponse.Status).JSON(validationResponse)
	}
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
	if validationResponse, err := utils.ValidateStruct(&payload); err != nil {
		return ctx.Status(validationResponse.Status).JSON(validationResponse)
	}

	ctxWithTimeout, cancel := withTimeout(ctx.Context(), defaultTimeout)
	defer cancel()

	response := c.petService.UpdatePetDetail(ctxWithTimeout, &payload)
	return ctx.Status(response.Status).JSON(response)
}

// @Summary Soft Delete Pet
// @Description soft delete pet by id
// @tags Pet
// @Security BearerAuth
// @Accept application/json
// @Produce application/json
// @Param pet_id path int true "Pet ID"
// @Success 200 {object} model.HTTPResponse
// @Failure 400 {object} model.HTTPResponse
// @Failure 401 {object} model.HTTPResponse
// @Failure 404 {object} model.HTTPResponse
// @Failure 500 {object} model.HTTPResponse
// @Router /pet/{pet_id} [delete]
func (c *petController) SoftDeletePet(ctx *fiber.Ctx) error {
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

	response := c.petService.SoftDeletePet(ctxWithTimeout, userID, uint(petID64))
	return ctx.Status(response.Status).JSON(response)
}
