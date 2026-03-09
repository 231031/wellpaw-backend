package controller

import (
	"net/http"
	"strconv"

	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/231031/wellpaw-backend/internal/service"
	"github.com/231031/wellpaw-backend/internal/utils"
	"github.com/gofiber/fiber/v2"
)

type PetCalendarController interface {
	CreatePetCalendar(ctx *fiber.Ctx) error
	GetPetCalendarsByPetID(ctx *fiber.Ctx) error
	GetPetCalendarsByUserID(ctx *fiber.Ctx) error
	GetCurrentMonthCalendarTypeSummaryByUserID(ctx *fiber.Ctx) error
}

type petCalendarController struct {
	petCalendarService service.PetCalendarService
}

func NewPetCalendarController(petCalendarService service.PetCalendarService) PetCalendarController {
	return &petCalendarController{
		petCalendarService: petCalendarService,
	}
}

// @Summary Create Pet Calendar
// @Description create a calendar and activity events for multiple pets
// @tags Pet Calendar
// @Security BearerAuth
// @Accept application/json
// @Produce application/json
// @Param CreatePetCalendarPaylaod body model.CreatePetCalendarPaylaod true "Create pet calendar payload"
// @Success 201 {object} model.CalendarResponse
// @Failure 400 {object} model.HTTPResponse
// @Failure 401 {object} model.HTTPResponse
// @Failure 404 {object} model.HTTPResponse
// @Failure 500 {object} model.HTTPResponse
// @Router /calendar [post]
func (c *petCalendarController) CreatePetCalendar(ctx *fiber.Ctx) error {
	userID := ctx.Locals("id").(uint)

	var payload model.CreatePetCalendarPaylaod
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

	response := c.petCalendarService.CreatePetCalendar(ctxWithTimeout, userID, &payload)
	return ctx.Status(response.Status).JSON(response)
}

// @Summary Get Pet Calendars By Pet ID
// @Description get pet calendars by pet id and current user ownership example of time 2025-03-10T21:00:00+07:00
// @tags Pet Calendar
// @Security BearerAuth
// @Accept application/json
// @Produce application/json
// @Param pet_id path int true "Pet ID"
// @Success 200 {object} model.ActivityCalendarsResponse
// @Failure 400 {object} model.HTTPResponse
// @Failure 401 {object} model.HTTPResponse
// @Failure 500 {object} model.HTTPResponse
// @Router /calendars/{pet_id} [get]
func (c *petCalendarController) GetPetCalendarsByPetID(ctx *fiber.Ctx) error {
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

	response := c.petCalendarService.GetPetCalendarsByPetID(ctxWithTimeout, userID, uint(petID64))
	return ctx.Status(response.Status).JSON(response)
}

// @Summary Get Pet Calendars By User ID
// @Description get all pet calendars of current user and preload pet
// @tags Pet Calendar
// @Security BearerAuth
// @Accept application/json
// @Produce application/json
// @Success 200 {object} model.CalendarsResponse
// @Failure 401 {object} model.HTTPResponse
// @Failure 500 {object} model.HTTPResponse
// @Router /calendars [get]
func (c *petCalendarController) GetPetCalendarsByUserID(ctx *fiber.Ctx) error {
	userID := ctx.Locals("id").(uint)

	ctxWithTimeout, cancel := withTimeout(ctx.Context(), defaultTimeout)
	defer cancel()

	response := c.petCalendarService.GetPetCalendarsByUserID(ctxWithTimeout, userID)
	return ctx.Status(response.Status).JSON(response)
}

// @Summary Get Current Month Pet Calendar Summary
// @Description get total calendars by type in current month of current user
// @tags Pet Calendar
// @Security BearerAuth
// @Accept application/json
// @Produce application/json
// @Success 200 {object} model.CalendarTypeSummaryResponse
// @Failure 401 {object} model.HTTPResponse
// @Failure 500 {object} model.HTTPResponse
// @Router /calendars/sum [get]
func (c *petCalendarController) GetCurrentMonthCalendarTypeSummaryByUserID(ctx *fiber.Ctx) error {
	userID := ctx.Locals("id").(uint)

	ctxWithTimeout, cancel := withTimeout(ctx.Context(), defaultTimeout)
	defer cancel()

	response := c.petCalendarService.GetCurrentMonthCalendarTypeSummaryByUserID(ctxWithTimeout, userID)
	return ctx.Status(response.Status).JSON(response)
}
