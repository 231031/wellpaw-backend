package service

import "github.com/231031/wellpaw-backend/internal/model"

type NutritientRequirementService interface {
	getNutritientFactor(pet model.PetType, rangeAge model.AgeType) (float64, float64)
	GetNutritientPerDay(rangeAge model.AgeType, p model.PetType, mer float64) (float64, float64)
}

type nutritientRequirementService struct {
	nutritientFactorByPet map[model.PetType]map[model.AgeType]map[string]float64
}

func NewNutritientRequirementService() NutritientRequirementService {
	service := &nutritientRequirementService{
		nutritientFactorByPet: map[model.PetType]map[model.AgeType]map[string]float64{
			model.CAT: {
				model.JUNIOR: {
					"protein": 75,
					"fat":     22.5,
				},
				model.ADULT: {
					"protein": 65,
					"fat":     22.5,
				},
				model.SENIOR: {
					"protein": 65,
					"fat":     22.5,
				},
			},
			model.DOG: {
				model.JUNIOR: {
					"protein": 56.3,
					"fat":     21.3,
				},
				model.ADULT: {
					"protein": 45,
					"fat":     13.8,
				},
				model.SENIOR: {
					"protein": 45,
					"fat":     13.8,
				},
			},
		},
	}
	return service
}

func (s *nutritientRequirementService) getNutritientFactor(pet model.PetType, rangeAge model.AgeType) (float64, float64) {
	ageMap, ok := s.nutritientFactorByPet[pet]
	if !ok {
		ageMap = s.nutritientFactorByPet[model.CAT]
	}
	nutMap, ok := ageMap[rangeAge]
	if !ok {
		nutMap = ageMap[model.ADULT]
	}
	return nutMap["protein"], nutMap["fat"]
}

func (s *nutritientRequirementService) GetNutritientPerDay(rangeAge model.AgeType, p model.PetType, mer float64) (float64, float64) {
	var pFactor, fFactor float64
	pFactor, fFactor = s.getNutritientFactor(p, rangeAge)
	return mer * (pFactor / 1000), mer * (fFactor / 1000)
}
