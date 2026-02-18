package service

import (
	"log"

	"github.com/231031/wellpaw-backend/internal/model"
)

type CalculationService interface {
	CalMerEnergyRequirement(petDetail *model.PetDetail, petType model.PetType) float64
	CalNutritientRequirement(mer float64, petDetail *model.PetDetail, petType model.PetType) (float64, float64)
	CalTotalIntakeFoodPlan(foodPlanDetails []*model.PetFoodPlanDetail) *model.PetFoodPlanTotal
	CalEnergyIntakeFromGramIntake(gramsIntake, energyFood float64, typeFood model.FoodType) float64
	CalNutritientIntakeFromGramIntake(gramsIntake, proteinFood, fatFood float64, typeFood model.FoodType) (float64, float64)
	calFeedingAmountEachFoodPerDay(energyIntake float64, food model.Food) *model.PetFoodPlanDetail
	CalFeedingAmountPerDay(petDetail *model.PetDetail, foods []model.Food) []*model.PetFoodPlanDetail
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

	proteinIntake, fatIntake := s.CalNutritientIntakeFromGramIntake(gramsIntake, food.Protein, food.Fat, food.Type)

	foodPlanDetail = &model.PetFoodPlanDetail{
		Amount:        gramsIntake,
		EnergyIntake:  energyIntake,
		ProteinIntake: proteinIntake,
		FatIntake:     fatIntake,
	}
	return foodPlanDetail
}

func (s *calculationService) CalFeedingAmountPerDay(petDetail *model.PetDetail, foods []model.Food) []*model.PetFoodPlanDetail {
	reqEnergy := petDetail.Energy

	checkType := map[model.FoodType]float64{}
	for _, f := range foods {
		checkType[f.Type] = f.Energy
	}

	threadHold := 0.2 * petDetail.Energy
	if supEnergy, ok := checkType[model.SUPPLEMENTS]; ok && supEnergy > threadHold {
		// energy per serving
		reqEnergy = reqEnergy - supEnergy
	}

	if _, ok := checkType[model.TREATS]; ok {
		checkType[model.TREATS] = 0.1

		foodPercent := 1.0 - checkType[model.TREATS]
		checkType[model.DRY] = foodPercent * 0.70
		checkType[model.WET] = foodPercent * 0.30
		log.Println("dry, wet, treat food")
	} else if _, ok := checkType[model.WET]; ok {
		checkType[model.DRY] = 0.70
		checkType[model.WET] = 0.30
		checkType[model.TREATS] = 0
		log.Println("dry and wet food")
	} else {
		checkType[model.DRY] = 1.0
		log.Println("just dry food")
	}

	var foodPlanDetails []*model.PetFoodPlanDetail
	for _, f := range foods {
		if f.Type == model.SUPPLEMENTS {
			foodPlanDetails = append(foodPlanDetails, &model.PetFoodPlanDetail{
				// feeding base on recommended of specific supplement
				Amount:        0,
				EnergyIntake:  f.Energy,
				ProteinIntake: f.Protein,
				FatIntake:     f.Fat,
			})
			continue
		}
		enegyIntakePerF := reqEnergy * checkType[f.Type]
		feedingDetail := s.calFeedingAmountEachFoodPerDay(enegyIntakePerF, f)
		foodPlanDetails = append(foodPlanDetails, feedingDetail)
	}
	return foodPlanDetails
}

func (s *calculationService) ConvertGramsToCup(foodPetFoodPlan *model.FoodPetFoodPlan) float64 {
	return 1.0
}

func (s *calculationService) CalExpectedWeight(currentWeight float64, bcs model.BcsType) float64 {
	return 1.0
}
