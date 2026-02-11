package service

import (
	"github.com/231031/wellpaw-backend/internal/model"
)

type CalculationService interface {
	ConvertNutricientPercentToGrams(proteinPercent, fatPercant, moisturePercent float64) (float64, float64)
	CalMerEnergyRequirement(petDetail *model.PetDetail, petType model.PetType) float64
	CalNutritientRequirement(mer float64, petDetail *model.PetDetail, petType model.PetType) (float64, float64)
	CalExpectedWeight(currentWeight float64, bcs model.BcsType) float64
}

type calculationService struct {
	energyRequirementService     EnergyRequirementService
	nutritientRequirementService NutritientRequirementService
}

func NewCalculationService(energyRequirementService EnergyRequirementService, nutritientRequirementService NutritientRequirementService) CalculationService {
	return &calculationService{
		energyRequirementService:     energyRequirementService,
		nutritientRequirementService: nutritientRequirementService,
	}
}

// not test
func (s *calculationService) ConvertNutricientPercentToGrams(proteinPercent, fatPercant, moisturePercent float64) (float64, float64) {
	dm := 100 - moisturePercent
	proteinGrams := (proteinPercent / 100) * dm
	fatGrams := (fatPercant / 100) * dm
	return proteinGrams, fatGrams
}

func (s *calculationService) CalMerEnergyRequirement(petDetail *model.PetDetail, petType model.PetType) float64 {
	return s.energyRequirementService.GetMerEnergy(
		petDetail.Weight,
		petDetail.AgeRange,
		petDetail.ActivityLevel,
		petDetail.BCS,
		petDetail.Gestation,
		petDetail.GestationStartDate,
		petDetail.Lactation,
		petDetail.Neutered,
		petType,
	)
}

func (s *calculationService) CalNutritientRequirement(mer float64, petDetail *model.PetDetail, petType model.PetType) (float64, float64) {
	return s.nutritientRequirementService.GetNutritientPerDay(
		petDetail.AgeRange,
		petType,
		mer,
	)
}

func (s *calculationService) CalAmountEachFoodPerDay(petDetail *model.PetDetail, food model.Food, propotion float64) float64 {
	return 0
}

func (s *calculationService) CalFoodsAmountPerDay(petDetail *model.PetDetail, foods []model.Food) []model.PetFoodPlanDetail {
	// check food type to get proper propotion in each case
	return nil
}

func (s *calculationService) CalExpectedWeight(currentWeight float64, bcs model.BcsType) float64 {
	return 1.0
}
