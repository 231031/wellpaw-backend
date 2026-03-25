package service

import (
	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/231031/wellpaw-backend/internal/utils"
)

type CalculationService interface {
	MapBcsScoreToBcsRange(bcsScore int) model.BcsType
	CalMerEnergyRequirement(petDetail *model.PetDetail, petType model.PetType) float64
	CalNutritientRequirement(mer float64, petDetail *model.PetDetail, petType model.PetType) (float64, float64)
	CalTotalIntakeFoodPlan(foodPlanDetails []*model.PetFoodPlanDetail) *model.PetFoodPlanTotal
	CalEnergyIntakeFromGramIntake(gramsIntake, energyFood float64, typeFood model.FoodType) float64
	CalNutritientIntakeFromGramIntake(gramsIntake, proteinFood, fatFood float64, typeFood model.FoodType) (float64, float64)
	calFeedingAmountEachFoodPerDay(energyIntake float64, food model.Food) *model.PetFoodPlanDetail
	CalFeedingAmountPerDay(petDetail *model.PetDetail, foods []model.Food, supAmount float64) []*model.PetFoodPlanDetail
	CalAvgPercentWeightChangePerMonth(monthlyDetails []model.PetMonthlyNutritionTWA, bcsScore int, petType model.PetType, ageRange model.AgeType) *model.AvgPercentWeightChangePerMonth
	ConvertGramsToCupInPlan(foodPlan *model.PetFoodPlan)
	CalculateGramsToCup(foodPetFoodPlan model.PetFoodPlanDetail) float64
}

type calculationService struct {
	energyRequirementService     EnergyRequirementService
	nutritientRequirementService NutritientRequirementService
	expectedWeightService        ExpectedWeightService
}

func NewCalculationService(energyRequirementService EnergyRequirementService, nutritientRequirementService NutritientRequirementService, exexpectedWeightService ExpectedWeightService) CalculationService {
	return &calculationService{
		energyRequirementService:     energyRequirementService,
		nutritientRequirementService: nutritientRequirementService,
		expectedWeightService:        exexpectedWeightService,
	}
}

func (s *calculationService) MapBcsScoreToBcsRange(bcsScore int) model.BcsType {
	switch bcsScore {
	case 1, 2:
		return model.VERYTHIN
	case 3, 4:
		return model.THIN
	case 5:
		return model.IDEAL
	case 6, 7:
		return model.OVERWEIGHT
	case 8, 9:
		return model.OBESITY
	default:
		return model.IDEAL
	}
}

func (s *calculationService) CalMerEnergyRequirement(petDetail *model.PetDetail, petType model.PetType) float64 {
	bcs := s.MapBcsScoreToBcsRange(petDetail.BCS)
	return s.energyRequirementService.GetMerEnergy(
		petDetail.Weight,
		petDetail.AgeRange,
		*petDetail.ActivityLevel,
		bcs,
		*petDetail.Gestation,
		petDetail.GestationStartDate,
		*petDetail.Lactation,
		*petDetail.Neutered,
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

func (s *calculationService) CalTotalIntakeFoodPlan(foodPlanDetails []*model.PetFoodPlanDetail) *model.PetFoodPlanTotal {
	var pTotal, fTotal, eTotal float64
	for _, p := range foodPlanDetails {
		eTotal += p.EnergyIntake
		pTotal += p.ProteinIntake
		fTotal += p.FatIntake
	}

	return &model.PetFoodPlanTotal{
		TotalEnergyIntake:  eTotal,
		TotalProteinIntake: pTotal,
		TotalFatIntake:     fTotal,
	}
}

// food with supplement type cal per recommend amount
func (s *calculationService) CalEnergyIntakeFromGramIntake(gramsIntake, energyFood float64, typeFood model.FoodType) float64 {
	if typeFood == model.SUPPLEMENTS {
		return gramsIntake * energyFood
	}

	calPerGram := energyFood / 100.0
	return gramsIntake * calPerGram
}

func (s *calculationService) CalNutritientIntakeFromGramIntake(gramsIntake, proteinFood, fatFood float64, typeFood model.FoodType) (float64, float64) {
	if typeFood == model.SUPPLEMENTS {
		return gramsIntake * proteinFood, gramsIntake * fatFood
	}

	proteinIntake := gramsIntake * (proteinFood / 100.0)
	fatIntake := gramsIntake * (fatFood / 100.0)
	return proteinIntake, fatIntake
}

func (s *calculationService) calFeedingAmountEachFoodPerDay(energyIntake float64, food model.Food) *model.PetFoodPlanDetail {
	var foodPlanDetail *model.PetFoodPlanDetail
	calPerGram := food.Energy / 100.0
	gramsIntake := energyIntake / calPerGram

	proteinIntake, fatIntake := s.CalNutritientIntakeFromGramIntake(gramsIntake, food.Protein, food.Fat, *food.Type)

	foodPlanDetail = &model.PetFoodPlanDetail{
		Amount:        gramsIntake,
		EnergyIntake:  energyIntake,
		ProteinIntake: proteinIntake,
		FatIntake:     fatIntake,
	}
	return foodPlanDetail
}

func (s *calculationService) CalFeedingAmountPerDay(petDetail *model.PetDetail, foods []model.Food, supAmount float64) []*model.PetFoodPlanDetail {
	reqEnergy := petDetail.Energy

	checkType := map[model.FoodType]float64{}
	for _, f := range foods {
		checkType[*f.Type] = f.Energy
	}

	if supEnergy, ok := checkType[model.SUPPLEMENTS]; ok {
		// energy per recommended serving
		reqEnergy = reqEnergy - supEnergy*supAmount
	}

	ratioByType := map[model.FoodType]float64{}
	_, hasDry := checkType[model.DRY]
	_, hasWet := checkType[model.WET]
	_, hasTreat := checkType[model.TREATS]

	if hasDry && hasWet && hasTreat {
		ratioByType[model.TREATS] = 0.1
		foodPercent := 1.0 - ratioByType[model.TREATS]
		ratioByType[model.DRY] = foodPercent * 0.70
		ratioByType[model.WET] = foodPercent * 0.30
	} else if hasDry && hasWet {
		ratioByType[model.DRY] = 0.70
		ratioByType[model.WET] = 0.30
	} else if hasDry && hasTreat {
		ratioByType[model.TREATS] = 0.10
		ratioByType[model.DRY] = 0.90
	} else if hasWet && hasTreat {
		ratioByType[model.TREATS] = 0.10
		ratioByType[model.WET] = 0.90
	} else if hasWet {
		ratioByType[model.WET] = 1.0
	} else {
		ratioByType[model.DRY] = 1.0
	}

	var foodPlanDetails []*model.PetFoodPlanDetail
	for _, f := range foods {
		if *f.Type == model.SUPPLEMENTS {
			foodPlanDetails = append(foodPlanDetails, &model.PetFoodPlanDetail{
				// feeding base on recommended of specific supplement (same unit intake with nutritient)
				Amount:        supAmount,
				EnergyIntake:  f.Energy * supAmount,
				ProteinIntake: f.Protein * supAmount,
				FatIntake:     f.Fat * supAmount,
			})
			continue
		}
		enegyIntakePerF := reqEnergy * ratioByType[*f.Type]
		feedingDetail := s.calFeedingAmountEachFoodPerDay(enegyIntakePerF, f)
		foodPlanDetails = append(foodPlanDetails, feedingDetail)
	}
	return foodPlanDetails
}

func (s *calculationService) CalculateGramsToCup(foodInPlan model.PetFoodPlanDetail) float64 {
	grams := foodInPlan.FoodPetFoodPlan.Food.GramsPerCup
	if grams == 0.0 {
		grams = 120.0
	}
	cupFeed := foodInPlan.Amount / grams
	return utils.RoundFloat(cupFeed, 2)
}

func (s *calculationService) ConvertGramsToCupInPlan(foodPlan *model.PetFoodPlan) {
	if len(foodPlan.PetFoodPlanTotals) > 0 {
		for idx, ft := range foodPlan.PetFoodPlanTotals[0].PetFoodPlanDetails {
			cup := s.CalculateGramsToCup(ft)
			foodPlan.PetFoodPlanTotals[0].PetFoodPlanDetails[idx].Cup = cup
		}

	}
}

func (s *calculationService) CalAvgPercentWeightChangePerMonth(monthlyDetails []model.PetMonthlyNutritionTWA, bcsScore int, petType model.PetType, ageRange model.AgeType) *model.AvgPercentWeightChangePerMonth {
	bcs := s.MapBcsScoreToBcsRange(bcsScore)
	avgWeight := s.expectedWeightService.GetAvgPercentWeightChangePerMonth(monthlyDetails, bcs, petType, ageRange)
	return avgWeight
}
