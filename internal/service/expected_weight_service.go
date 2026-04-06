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
				model.SENIOR: {4, 8},
			},
			model.CAT: {
				model.JUNIOR: {20, 40},
				model.ADULT:  {2, 4},
				model.SENIOR: {2, 4},
			},
		},
		lossWeightFactor: map[model.PetType]map[model.AgeType][2]float64{
			model.DOG: {
				model.ADULT:  {4, 8},
				model.SENIOR: {4, 8},
			},
			model.CAT: {
				model.ADULT:  {2, 4},
				model.SENIOR: {2, 4},
			},
		},
	}
}

func (s *expectedWeightService) GetAvgPercentWeightChangePerMonth(monthlyDetails []model.PetMonthlyNutritionTWA, bcs model.BcsType, petType model.PetType, ageRange model.AgeType) *model.AvgPercentWeightChangePerMonth {
	if bcs == model.IDEAL {
		return &model.AvgPercentWeightChangePerMonth{
			Message: "คะแนนสภาพร่างกายอยู่ในเกณฑ์ปกติ รักษาน้ำหนักต่อไป",
		}
	}

	isGainCase := bcs == model.VERYTHIN || bcs == model.THIN
	isLossCase := bcs == model.OVERWEIGHT || bcs == model.OBESITY

	if ageRange == model.JUNIOR && isLossCase {
		return &model.AvgPercentWeightChangePerMonth{
			Message: "คะแนนสภาพร่างกายอยู่ในเกณฑ์อ้วนเกินไปสำหรับสัตว์เลี้ยงวัยเด็ก ควรปรึกษาสัตวแพทย์เพื่อประเมินสุขภาพโดยรวมและแผนการให้อาหารที่เหมาะสม",
		}
	}

	calculatedPercent, totalMonthsUsed := s.calculateAvgPercentFromMonthlyWeights(monthlyDetails)
	result := &model.AvgPercentWeightChangePerMonth{
		CalculatedPercent: calculatedPercent,
		TotalMonthsUsed:   totalMonthsUsed,
	}

	if isGainCase {
		infoMsg := "คะแนนสภาพร่างกายอยู่ในเกณฑ์ผอมเกินไป ควรเพิ่มน้ำหนักอย่างค่อยเป็นค่อยไปเพื่อสุขภาพที่ดีขึ้น"
		if ageRange == model.SENIOR {
			infoMsg = "ติดตามการเปลี่ยนแปลงน้ำหนักอย่างใกล้ชิด"
		}

		if startEnd, ok := s.gainWeightFactor[petType][ageRange]; ok {
			result.StartPercent = startEnd[0]
			result.EndPercent = startEnd[1]
			result.Message = infoMsg
			return result
		}

		return &model.AvgPercentWeightChangePerMonth{
			CalculatedPercent: result.CalculatedPercent,
			TotalMonthsUsed:   result.TotalMonthsUsed,
			Message:           infoMsg,
		}
	}

	if isLossCase {
		infoMsg := "คะแนนสภาพร่างกายอยู่ในเกณฑ์อ้วนเกินไป ควรลดน้ำหนักอย่างค่อยเป็นค่อยไปเพื่อสุขภาพที่ดีขึ้น"
		if ageRange == model.SENIOR {
			infoMsg = "ติดตามการเปลี่ยนแปลงน้ำหนักอย่างใกล้ชิด"
		}

		if startEnd, ok := s.lossWeightFactor[petType][ageRange]; ok {
			// Keep start <= end while representing weight loss as negative percent.
			result.StartPercent = -startEnd[1]
			result.EndPercent = -startEnd[0]
			result.Message = infoMsg
			return result
		}

		return &model.AvgPercentWeightChangePerMonth{
			CalculatedPercent: result.CalculatedPercent,
			TotalMonthsUsed:   result.TotalMonthsUsed,
			Message:           infoMsg,
		}
	}

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
