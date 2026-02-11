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

type PetService interface {
	CreateNewPet(ctx context.Context, pet *model.PetPayload) *model.HTTPResponse
	UpdatePetInfo(ctx context.Context, petInfo *model.Pet) *model.HTTPResponse
	UpdatePetDetail(ctx context.Context, petDetail *model.PetDetail) *model.HTTPResponse
}

type petService struct {
	calculationService CalculationService
	petRepo            repository.PetRepository
}

func NewPetService(calculationService CalculationService, petRepo repository.PetRepository) PetService {
	return &petService{
		calculationService: calculationService,
		petRepo:            petRepo,
	}
}

// not test
func (s *petService) GetAgeRangeFromBirthDate(petType model.PetType, birthDate time.Time) model.AgeType {
	ageMonth := time.Since(birthDate).Hours() / 24 / 30
	switch petType {
	case model.CAT:
		if ageMonth < 12 {
			return model.JUNIOR
		} else if ageMonth >= 84 {
			return model.SENIOR
		} else {
			return model.ADULT
		}
	case model.DOG:
		if ageMonth < 12 {
			return model.JUNIOR
		} else if ageMonth >= 84 {
			return model.SENIOR
		} else {
			return model.ADULT
		}
	}
	return model.ADULT
}

func (s *petService) CreateNewPet(ctx context.Context, pet *model.PetPayload) *model.HTTPResponse {
	petInfo := pet.PetInfo
	petDetail := pet.PetDetail

	petDetail.AgeRange = s.GetAgeRangeFromBirthDate(petInfo.Type, petInfo.BirthDate)
	petDetail.Energy = s.calculationService.CalMerEnergyRequirement(&petDetail, petInfo.Type)
	petDetail.Protein, petDetail.Fat = s.calculationService.CalNutritientRequirement(petDetail.Energy, &petDetail, petInfo.Type)

	petDetail.ExpectedWeight = s.calculationService.CalExpectedWeight(petDetail.Weight, petDetail.BCS)
	err := s.petRepo.CreateNewPet(ctx, &petInfo, &petDetail)
	if err != nil {
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToCreateMsg + "new pet",
		}
	}

	return &model.HTTPResponse{
		Status: http.StatusCreated,
		Data: map[string]interface{}{
			"pet_info":   petInfo,
			"pet_detail": petDetail,
		},
	}
}

func (s *petService) UpdatePetInfo(ctx context.Context, petInfo *model.Pet) *model.HTTPResponse {
	if err := s.petRepo.UpdatePetInfo(ctx, petInfo); err != nil {
		if errors.Is(err, utils.ErrNoRowsUpdated) {
			return &model.HTTPResponse{
				Status:  http.StatusNotFound,
				Message: "pet" + utils.NotFoundMsg,
			}
		}

		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToUpdateMsg + "pet info",
		}
	}

	petInfo, err := s.petRepo.GetPetInfoByID(ctx, petInfo.ID)
	if err != nil {
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: "pet information is updated, but failed to get new data",
		}
	}
	return &model.HTTPResponse{
		Status: http.StatusOK,
		Data:   petInfo,
	}
}

func (s *petService) UpdatePetDetail(ctx context.Context, petDetail *model.PetDetail) *model.HTTPResponse {
	pet, err := s.petRepo.GetPetInfoByID(ctx, petDetail.PetID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &model.HTTPResponse{
				Status:  http.StatusNotFound,
				Message: "pet" + utils.NotFoundMsg,
			}
		}

		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToGetMsg + "pet",
		}
	}

	petDetail.Energy = s.calculationService.CalMerEnergyRequirement(petDetail, pet.Type)
	petDetail.Protein, petDetail.Fat = s.calculationService.CalNutritientRequirement(petDetail.Energy, petDetail, pet.Type)
	petDetail.ExpectedWeight = s.calculationService.CalExpectedWeight(petDetail.Weight, petDetail.BCS)

	// if the some feeding plan is active -> calculate new feeding amount for this updating
	// transaction update instead of one

	if err := s.petRepo.UpdatePetDetails(ctx, petDetail); err != nil {
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToUpdateMsg + "pet detail",
		}
	}

	return &model.HTTPResponse{
		Status: http.StatusOK,
		Data: map[string]interface{}{
			"pet_info":   pet,
			"pet_detail": petDetail,
		},
	}
}
