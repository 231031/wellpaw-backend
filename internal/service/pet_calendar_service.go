package service

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/231031/wellpaw-backend/internal/repository"
	"github.com/231031/wellpaw-backend/internal/utils"
	"gorm.io/gorm"
)

type PetCalendarService interface {
	CreatePetCalendar(ctx context.Context, userID uint, payload *model.CreatePetCalendarPaylaod) *model.HTTPResponse
	GetPetCalendarsByPetID(ctx context.Context, userID uint, petID uint) *model.HTTPResponse
	GetPetCalendarsByUserID(ctx context.Context, userID uint) *model.HTTPResponse
	GetCurrentMonthCalendarTypeSummaryByUserID(ctx context.Context, userID uint) *model.HTTPResponse
}

type petCalendarService struct {
	petCalendarRepo repository.PetCalendarRepository
}

func NewPetCalendarService(petCalendarRepo repository.PetCalendarRepository) PetCalendarService {
	return &petCalendarService{
		petCalendarRepo: petCalendarRepo,
	}
}

func (s *petCalendarService) CreatePetCalendar(ctx context.Context, userID uint, payload *model.CreatePetCalendarPaylaod) *model.HTTPResponse {
	petIDs, invalidPayloadResp := s.validatePetIDs(payload.PetIDs)
	if invalidPayloadResp != nil {
		return invalidPayloadResp
	}

	petCalendar := &model.PetCalendar{
		Name:          payload.PetCalendar.Name,
		Type:          payload.PetCalendar.Type,
		StartDatetime: payload.PetCalendar.StartDatetime,
		EndDate:       payload.PetCalendar.EndDate,
		Frequently:    payload.PetCalendar.Frequently,
		Notation:      payload.PetCalendar.Notation,
	}

	if !petCalendar.EndDate.IsZero() && petCalendar.EndDate.Before(petCalendar.StartDatetime) {
		return &model.HTTPResponse{
			Status:  http.StatusBadRequest,
			Message: "end date must be greater than or equal to start datetime",
		}
	}

	err := s.petCalendarRepo.CreatePetCalendar(ctx, userID, petCalendar, petIDs)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &model.HTTPResponse{
				Status:  http.StatusNotFound,
				Message: "some pets" + utils.NotFoundMsg,
			}
		}

		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToCreateMsg + "pet calendar",
		}
	}

	responseCalendar := s.mapPetCalendarToThaiTimezone(*petCalendar)
	return &model.HTTPResponse{
		Status: http.StatusCreated,
		Data:   responseCalendar,
	}
}

func (s *petCalendarService) GetPetCalendarsByPetID(ctx context.Context, userID uint, petID uint) *model.HTTPResponse {
	activityEvents, err := s.petCalendarRepo.GetPetCalendarsByPetIDAndUserID(ctx, petID, userID)
	if err != nil {
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToGetMsg + "pet calendars",
		}
	}

	responseActivityEvents := s.mapPetActivityEventsToThaiTimezone(activityEvents)
	return &model.HTTPResponse{
		Status: http.StatusOK,
		Data: map[string]interface{}{
			"activity_calendars": responseActivityEvents,
		},
	}
}

func (s *petCalendarService) GetPetCalendarsByUserID(ctx context.Context, userID uint) *model.HTTPResponse {
	petCalendars, err := s.petCalendarRepo.GetPetCalendarsByUserID(ctx, userID)
	if err != nil {
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToGetMsg + "pet calendars",
		}
	}

	responseCalendars := s.mapPetCalendarsToThaiTimezone(petCalendars)
	return &model.HTTPResponse{
		Status: http.StatusOK,
		Data: map[string]interface{}{
			"calendars": responseCalendars,
		},
	}
}

func (s *petCalendarService) GetCurrentMonthCalendarTypeSummaryByUserID(ctx context.Context, userID uint) *model.HTTPResponse {
	monthStart, nextMonthStart := utils.GetMonthRangeInThai(time.Now())
	typeCountMap, err := s.petCalendarRepo.GetCurrentMonthCalendarTypeCountByUserID(ctx, userID, monthStart, nextMonthStart)
	if err != nil {
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToGetMsg + "pet calendar summary",
		}
	}

	summary := []model.CalendarTypeSummary{
		{
			CalendarType: model.VACCINE,
			Type:         model.VACCINE.String(),
			Times:        int(typeCountMap[model.VACCINE]),
		},
		{
			CalendarType: model.DRUG,
			Type:         model.DRUG.String(),
			Times:        int(typeCountMap[model.DRUG]),
		},
		{
			CalendarType: model.APPOINTMENT,
			Type:         model.APPOINTMENT.String(),
			Times:        int(typeCountMap[model.APPOINTMENT]),
		},
	}

	return &model.HTTPResponse{
		Status: http.StatusOK,
		Data:   summary,
	}
}

func (s *petCalendarService) validatePetIDs(petIDs []uint) ([]uint, *model.HTTPResponse) {
	seen := make(map[uint]struct{}, len(petIDs))
	normalizedPetIDs := make([]uint, 0, len(petIDs))

	for _, petID := range petIDs {
		if petID == 0 {
			return nil, &model.HTTPResponse{
				Status:  http.StatusBadRequest,
				Message: "invalid pet",
			}
		}

		if _, exists := seen[petID]; exists {
			return nil, &model.HTTPResponse{
				Status:  http.StatusBadRequest,
				Message: "duplicate pet is not allowed",
			}
		}

		seen[petID] = struct{}{}
		normalizedPetIDs = append(normalizedPetIDs, petID)
	}

	return normalizedPetIDs, nil
}

func (s *petCalendarService) mapPetActivityEventsToThaiTimezone(activityEvents []model.PetActivityCalendar) []model.PetActivityCalendar {
	if len(activityEvents) == 0 {
		return []model.PetActivityCalendar{}
	}

	for idx := range activityEvents {
		if activityEvents[idx].PetCalendar != nil {
			s.convertPetCalendarToThaiTimezone(activityEvents[idx].PetCalendar)
		}

		if activityEvents[idx].Pet != nil {
			activityEvents[idx].Pet.BirthDate = utils.ConvertTimeToThaiTimezone(activityEvents[idx].Pet.BirthDate)
			activityEvents[idx].Pet.CreatedAt = utils.ConvertTimeToThaiTimezone(activityEvents[idx].Pet.CreatedAt)
			activityEvents[idx].Pet.UpdatedAt = utils.ConvertTimeToThaiTimezone(activityEvents[idx].Pet.UpdatedAt)
		}
	}

	return activityEvents
}

func (s *petCalendarService) mapPetCalendarsToThaiTimezone(petCalendars []model.PetCalendar) []model.PetCalendar {
	if len(petCalendars) == 0 {
		return []model.PetCalendar{}
	}

	for idx := range petCalendars {
		s.convertPetCalendarToThaiTimezone(&petCalendars[idx])

		if len(petCalendars[idx].ActivityEvents) == 0 {
			petCalendars[idx].ActivityEvents = []model.PetActivityCalendar{}
			continue
		}

		for eventIdx := range petCalendars[idx].ActivityEvents {
			if petCalendars[idx].ActivityEvents[eventIdx].Pet == nil {
				continue
			}

			petCalendars[idx].ActivityEvents[eventIdx].Pet.BirthDate = utils.ConvertTimeToThaiTimezone(petCalendars[idx].ActivityEvents[eventIdx].Pet.BirthDate)
			petCalendars[idx].ActivityEvents[eventIdx].Pet.CreatedAt = utils.ConvertTimeToThaiTimezone(petCalendars[idx].ActivityEvents[eventIdx].Pet.CreatedAt)
			petCalendars[idx].ActivityEvents[eventIdx].Pet.UpdatedAt = utils.ConvertTimeToThaiTimezone(petCalendars[idx].ActivityEvents[eventIdx].Pet.UpdatedAt)
		}
	}

	return petCalendars
}

func (s *petCalendarService) mapPetCalendarToThaiTimezone(petCalendar model.PetCalendar) model.PetCalendar {
	s.convertPetCalendarToThaiTimezone(&petCalendar)
	return petCalendar
}

func (s *petCalendarService) convertPetCalendarToThaiTimezone(petCalendar *model.PetCalendar) {
	if petCalendar == nil {
		return
	}

	petCalendar.StartDatetime = utils.ConvertTimeToThaiTimezone(petCalendar.StartDatetime)
	petCalendar.EndDate = utils.ConvertTimeToThaiTimezone(petCalendar.EndDate)
	petCalendar.CreatedAt = utils.ConvertTimeToThaiTimezone(petCalendar.CreatedAt)
}
