package service

import (
	"context"
	"time"

	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/231031/wellpaw-backend/internal/repository"
)

type PetService interface {
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

func (s *petService) CreateNewPet(ctx context.Context, pet *model.CreatePetPayload) error {
	petInfo := pet.PetInfo
	petDetail := pet.PetDetail

	petDetail.AgeRange = s.GetAgeRangeFromBirthDate(petInfo.Type, petInfo.BirthDate)
	petDetail.Energy = s.calculationService.CalMerEnergyRequirement(&petDetail, petInfo.Type)
	petDetail.Protein, petDetail.Fat = s.calculationService.CalNutritientRequirement(petDetail.Energy, &petDetail, petInfo.Type)

	err := s.petRepo.CreateNewPet(ctx, &petInfo, &petDetail)
	if err != nil {
		return err
	}
	return nil
}

func (s *petService) UpdatePetDetail(ctx context.Context, petDetail *model.PetDetail) error {
	// calculate new feeding quantity of active plan based on new pet detail
	// transaction create new feeding quantity base on active plan and create new pet detail
	return nil
}
