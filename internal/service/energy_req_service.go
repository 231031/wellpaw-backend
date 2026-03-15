package service

import (
	"math"
	"time"

	"github.com/231031/wellpaw-backend/internal/applogger"
	"github.com/231031/wellpaw-backend/internal/model"
)

type EnergyRequirementService interface {
	getNeuteredFactor(neutered bool, p model.PetType) float64
	getActivityBcsFactor(al model.ActivityLevel, bcs model.BcsType) (float64, float64)
	RERJuniorCat(bw float64) float64
	RERAdultCat(bw float64) float64
	RERJuniorDog(bw float64) float64
	RERAdultDog(bw float64) float64
	getEnergyFormula(pet model.PetType, age model.AgeType) RerOperation
	GetMerEnergy(bw float64, rangeAge model.AgeType, al model.ActivityLevel, bcs model.BcsType, gestation bool, gestationDate time.Time, lactation bool, neutered bool, pet model.PetType) float64
}

type energyRequirementService struct {
	stageThreeDays        int
	neuteredFactorByPet   map[model.PetType]map[bool]float64
	activityFactorByLevel map[model.ActivityLevel]float64
	bcsFactorByType       map[model.BcsType]float64
	rerByPetAge           map[model.PetType]map[model.AgeType]RerOperation
}

func NewEnergyRequirementService() EnergyRequirementService {
	service := &energyRequirementService{
		stageThreeDays: 42,
		neuteredFactorByPet: map[model.PetType]map[bool]float64{
			model.CAT: {
				false: 1.4,
				true:  1.2,
			},
			model.DOG: {
				false: 1.8,
				true:  1.6,
			},
		},
		activityFactorByLevel: map[model.ActivityLevel]float64{
			model.INACTIVE:   1,
			model.SOMEACTIVE: 1.2,
			model.ACTIVE:     1.4,
			model.VERYACTIVE: 1.6,
		},
		bcsFactorByType: map[model.BcsType]float64{
			model.VERYTHIN:   1.2,
			model.THIN:       1.2,
			model.IDEAL:      1,
			model.OVERWEIGHT: 0.8,
			model.OBESITY:    0.8,
		},
	}

	service.rerByPetAge = map[model.PetType]map[model.AgeType]RerOperation{
		model.CAT: {
			model.JUNIOR: service.RERJuniorCat,
			model.ADULT:  service.RERAdultCat,
			model.SENIOR: service.RERAdultCat,
		},
		model.DOG: {
			model.JUNIOR: service.RERJuniorDog,
			model.ADULT:  service.RERAdultDog,
			model.SENIOR: service.RERAdultDog,
		},
	}
	return service
}

type RerOperation func(bw float64) float64

func (s *energyRequirementService) getNeuteredFactor(neutered bool, p model.PetType) float64 {
	if byNeutered, ok := s.neuteredFactorByPet[p]; ok {
		if factor, ok := byNeutered[neutered]; ok {
			return factor
		}
	}
	return 1.0
}

func (s *energyRequirementService) getActivityBcsFactor(al model.ActivityLevel, bcs model.BcsType) (float64, float64) {
	alFactor := 1.0
	if factor, ok := s.activityFactorByLevel[al]; ok {
		alFactor = factor
	}

	bcsFactor := 1.0
	if factor, ok := s.bcsFactorByType[bcs]; ok {
		bcsFactor = factor
	}

	return alFactor, bcsFactor
}

func (s *energyRequirementService) rer(bw float64, exponent float64, multiplier float64) float64 {
	return 70 * math.Pow(bw, exponent) * multiplier
}

func (s *energyRequirementService) RERJuniorCat(bw float64) float64 {
	return s.rer(bw, 0.67, 2)
}

func (s *energyRequirementService) RERAdultCat(bw float64) float64 {
	return s.rer(bw, 0.67, 1)
}

func (s *energyRequirementService) RERJuniorDog(bw float64) float64 {
	return s.rer(bw, 0.75, 2)
}

func (s *energyRequirementService) RERAdultDog(bw float64) float64 {
	return s.rer(bw, 0.75, 1)
}

func (s *energyRequirementService) getEnergyFormula(pet model.PetType, age model.AgeType) RerOperation {
	ageMap, ok := s.rerByPetAge[pet]
	if !ok {
		applogger.LogError("failed to find RER formula for pet type", serviceLog)
		ageMap = s.rerByPetAge[model.CAT]
	}
	if formula, ok := ageMap[age]; ok {
		return formula
	}
	return ageMap[model.ADULT]
}

func (s *energyRequirementService) GetMerEnergy(bw float64, rangeAge model.AgeType, al model.ActivityLevel, bcs model.BcsType, gestation bool, gestationDate time.Time, lactation bool, neutered bool, pet model.PetType) float64 {
	rerFormula := s.getEnergyFormula(pet, rangeAge)
	if rangeAge == model.JUNIOR {
		return rerFormula(bw)
	}

	alFactor, bcsFactor := s.getActivityBcsFactor(al, bcs)
	reproductionFactor := s.getNeuteredFactor(neutered, pet)

	return rerFormula(bw) * alFactor * bcsFactor * reproductionFactor
}
