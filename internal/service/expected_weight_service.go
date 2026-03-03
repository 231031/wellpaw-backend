package service

import (
	"github.com/231031/wellpaw-backend/internal/model"
)

type ExpectedWeightService interface {
	GetAvgPercentWeightChangePerMonth(monthlyDetails []model.PetMonthlyNutritionTWA, bcs model.BcsType, petType model.PetType, ageRange model.AgeType) *model.AvgPercentWeightChangePerMonth
}

type expectedWeightService struct {
	gainWeightFactor map[model.PetType]map[model.AgeType][2]float64
	lossWeightFactor map[model.PetType]map[model.AgeType][2]float64
}

func NewExpectedWeightService() ExpectedWeightService {
	return &expectedWeightService{
		gainWeightFactor: map[model.PetType]map[model.AgeType][2]float64{
			model.DOG: {
				model.JUNIOR: {20, 40},
				model.ADULT:  {4, 8},
			},
			model.CAT: {
				model.JUNIOR: {20, 40},
				model.ADULT:  {2, 4},
			},
		},
		lossWeightFactor: map[model.PetType]map[model.AgeType][2]float64{
			model.DOG: {
				model.ADULT: {4, 8},
			},
			model.CAT: {
				model.ADULT: {2, 4},
			},
		},
	}
}

func (s *expectedWeightService) GetAvgPercentWeightChangePerMonth(monthlyDetails []model.PetMonthlyNutritionTWA, bcs model.BcsType, petType model.PetType, ageRange model.AgeType) *model.AvgPercentWeightChangePerMonth {
	infoMsg := "the information isn't provided for this condition"

	if ageRange == model.SENIOR {
		return &model.AvgPercentWeightChangePerMonth{
			Message: "monitor the weight closely",
		}
	}

	isGainCase := bcs == model.VERYTHIN || bcs == model.THIN
	isLossCase := bcs == model.OVERWEIGHT || bcs == model.OBESITY

	if ageRange == model.JUNIOR && isLossCase {
		return &model.AvgPercentWeightChangePerMonth{
			Message: infoMsg,
		}
	}

	calculatedPercent, totalMonthsUsed := s.calculateAvgPercentFromMonthlyWeights(monthlyDetails)
	result := &model.AvgPercentWeightChangePerMonth{
		CalculatedPercent: calculatedPercent,
		TotalMonthsUsed:   totalMonthsUsed,
	}

	if isGainCase {
		if startEnd, ok := s.gainWeightFactor[petType][ageRange]; ok {
			result.StartPercent = startEnd[0]
			result.EndPercent = startEnd[1]
			return result
		}

		return &model.AvgPercentWeightChangePerMonth{
			CalculatedPercent: result.CalculatedPercent,
			TotalMonthsUsed:   result.TotalMonthsUsed,
			Message:           infoMsg,
		}
	}

	if isLossCase {
		if startEnd, ok := s.lossWeightFactor[petType][ageRange]; ok {
			// Keep start <= end while representing weight loss as negative percent.
			result.StartPercent = -startEnd[1]
			result.EndPercent = -startEnd[0]
			return result
		}

		return &model.AvgPercentWeightChangePerMonth{
			CalculatedPercent: result.CalculatedPercent,
			TotalMonthsUsed:   result.TotalMonthsUsed,
			Message:           infoMsg,
		}
	}

	// IDEAL BCS keeps range at 0 with no message.
	return result
}

func (s *expectedWeightService) calculateAvgPercentFromMonthlyWeights(monthlyDetails []model.PetMonthlyNutritionTWA) (float64, int) {
	if len(monthlyDetails) < 2 {
		return 0, 0
	}
	var totalPercent float64
	totalMonths := 0

	for i := 1; i < len(monthlyDetails); i++ {
		prev := monthlyDetails[i-1].Weight
		curr := monthlyDetails[i].Weight
		if prev <= 0 {
			continue
		}

		totalPercent += ((curr - prev) / prev) * 100
		totalMonths += 1
	}

	if totalMonths == 0 {
		return 0, 0
	}

	return totalPercent / float64(totalMonths), totalMonths
}
