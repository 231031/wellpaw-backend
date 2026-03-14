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

type PetFoodPlanService interface {
	getInfoForPetFoodPlan(ctx context.Context, userID uint, petID uint, foodIDs []uint) (*model.PetDetail, []model.Food, *model.HTTPResponse)
	CalculatePetFoodPlan(ctx context.Context, userID uint, payload *model.CalculatePetFoodPlanPayload) *model.HTTPResponse
	CreatePetFoodPlan(ctx context.Context, userID uint, payload *model.CreatePetFoodPlanPayload) *model.HTTPResponse
	GetLastestActivePlanDetailByPet(ctx context.Context, petID uint) *model.HTTPResponse
	UpdateFeedingAmountFromUser(ctx context.Context, payload *model.AdjustAmountFoodInPetFoodPlanPayload) *model.HTTPResponse
}

type petFoodPlanService struct {
	calculationService    CalculationService
	petFoodPlanRepo       repository.PetFoodPlanRepository
	petRepo               repository.PetRepository
	foodRepo              repository.FoodRepository
	freeValidationService FreeTierUsageValidationService
}

func NewPetFoodPlanService(calculationService CalculationService, petFoodPlanRepo repository.PetFoodPlanRepository, petRepo repository.PetRepository, foodRepo repository.FoodRepository, freeTierUsageValidationService FreeTierUsageValidationService) PetFoodPlanService {
	return &petFoodPlanService{
		calculationService:    calculationService,
		petFoodPlanRepo:       petFoodPlanRepo,
		petRepo:               petRepo,
		foodRepo:              foodRepo,
		freeValidationService: freeTierUsageValidationService,
	}
}

func (s *petFoodPlanService) getInfoForPetFoodPlan(ctx context.Context, userID uint, petID uint, foodIDs []uint) (*model.PetDetail, []model.Food, *model.HTTPResponse) {
	petDetail, err := s.petRepo.GetLatestPetDetailByPetID(ctx, petID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, &model.HTTPResponse{
				Status:  http.StatusNotFound,
				Message: "pet detail" + utils.NotFoundMsg,
			}
		}

		return nil, nil, &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToGetMsg + "pet detail",
		}
	}

	foods, err := s.foodRepo.GetFoodsByIDsAndUserID(ctx, userID, foodIDs)
	if err != nil {
		return nil, nil, &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToGetMsg + "foods",
		}
	}

	if len(foods) != len(foodIDs) {
		return nil, nil, &model.HTTPResponse{
			Status:  http.StatusNotFound,
			Message: "some foods" + utils.NotFoundMsg,
		}
	}

	return petDetail, foods, nil
}

func (s *petFoodPlanService) CalculatePetFoodPlan(ctx context.Context, userID uint, payload *model.CalculatePetFoodPlanPayload) *model.HTTPResponse {
	foodIDs := make([]uint, 0, len(payload.Foods))
	seenFoodIDs := make(map[uint]bool, len(payload.Foods))
	gramsPerCupByFoodID := make(map[uint]float64, len(payload.Foods))

	for _, f := range payload.Foods {
		if _, exists := seenFoodIDs[f.FoodID]; exists {
			return &model.HTTPResponse{
				Status:  http.StatusBadRequest,
				Message: "duplicate food_id is not allowed",
			}
		}
		seenFoodIDs[f.FoodID] = true

		if *payload.Unit == model.CUP && f.GramsPerCup == nil {
			return &model.HTTPResponse{
				Status:  http.StatusBadRequest,
				Message: "grams per cup is required for unit cup",
			}
		}

		if f.GramsPerCup != nil {
			if *f.GramsPerCup <= 0.0 {
				return &model.HTTPResponse{
					Status:  http.StatusBadRequest,
					Message: "grams per cup must more than 0",
				}
			}
			gramsPerCupByFoodID[f.FoodID] = *f.GramsPerCup
		}

		foodIDs = append(foodIDs, f.FoodID)
	}

	petDetail, foods, resp := s.getInfoForPetFoodPlan(ctx, userID, payload.PetID, foodIDs)
	if resp != nil {
		return resp
	}

	foodPlanDetails := make([]*model.PetFoodPlanDetail, 0, len(payload.Foods))
	if payload.Foods[0].Amount > 0 {
		foodsDetail := map[uint]model.Food{}
		for _, food := range foods {
			foodsDetail[food.ID] = food
		}

		for _, food := range payload.Foods {
			if food.Amount <= 0 {
				return &model.HTTPResponse{
					Status:  http.StatusBadRequest,
					Message: "invalid amount of food",
				}
			}
			foodDetail := foodsDetail[food.FoodID]
			energy := s.calculationService.CalEnergyIntakeFromGramIntake(food.Amount, foodDetail.Energy, *foodDetail.Type)
			protein, fat := s.calculationService.CalNutritientIntakeFromGramIntake(food.Amount, foodDetail.Protein, foodDetail.Fat, *foodDetail.Type)
			foodPlanDetails = append(foodPlanDetails, &model.PetFoodPlanDetail{
				Amount:        food.Amount,
				EnergyIntake:  energy,
				ProteinIntake: protein,
				FatIntake:     fat,
			})
		}
	} else {
		foodPlanDetails = s.calculationService.CalFeedingAmountPerDay(petDetail, foods)
	}

	if len(foodPlanDetails) != len(foods) {
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: "failed to calculate feeding plan details",
		}
	}

	for idx := range foods {
		foodPlanDetails[idx].FoodPetFoodPlanID = foods[idx].ID
	}

	plan := &model.PetFoodPlan{
		PetID:     payload.PetID,
		Name:      payload.Name,
		Active:    false,
		Unit:      *payload.Unit,
		CreatedAt: time.Now(),
	}

	foodsInPlan := make([]model.FoodPetFoodPlan, 0, len(foods))
	for _, food := range payload.Foods {
		gramsPerCup := 0.0
		if *payload.Unit == model.CUP {
			gramsPerCup = gramsPerCupByFoodID[food.FoodID]
		}

		foodsInPlan = append(foodsInPlan, model.FoodPetFoodPlan{
			FoodID:      food.FoodID,
			GramsPerCup: gramsPerCup,
		})

	}

	foodPlanTotal := s.calculationService.CalTotalIntakeFoodPlan(foodPlanDetails)
	foodPlanTotal.PetDetailID = petDetail.ID
	foodPlanTotal.PetDetail = petDetail

	plan.PetFoodPlanTotals = append(plan.PetFoodPlanTotals, *foodPlanTotal)
	for _, detail := range foodPlanDetails {
		if detail == nil {
			continue
		}
		plan.PetFoodPlanTotals[0].PetFoodPlanDetails = append(plan.PetFoodPlanTotals[0].PetFoodPlanDetails, *detail)
	}

	plan.FoodPetFoodPlans = foodsInPlan
	for idx, fp := range plan.FoodPetFoodPlans {
		plan.PetFoodPlanTotals[0].PetFoodPlanDetails[idx].FoodPetFoodPlan = &fp
	}

	if plan.Unit == model.CUP {
		s.calculationService.ConvertGramsToCupInPlan(plan)
	}
	return &model.HTTPResponse{
		Status: http.StatusOK,
		Data: map[string]interface{}{
			"pet_food_plan": plan,
		},
	}
}

func (s *petFoodPlanService) CreatePetFoodPlan(ctx context.Context, userID uint, payload *model.CreatePetFoodPlanPayload) *model.HTTPResponse {
	tier, freeUsage, resp := s.freeValidationService.CheckValidUsageByUserID(ctx, userID, model.FOODPLAN)
	if resp != nil {
		return resp
	}

	foodIDs := make([]uint, 0, len(payload.Foods))
	seenFoodIDs := make(map[uint]bool, len(payload.Foods))
	gramsPerCupByFoodID := make(map[uint]float64, len(payload.Foods))

	for _, f := range payload.Foods {
		if _, exists := seenFoodIDs[f.FoodID]; exists {
			return &model.HTTPResponse{
				Status:  http.StatusBadRequest,
				Message: "duplicate food_id is not allowed",
			}
		}
		seenFoodIDs[f.FoodID] = true

		if *payload.Unit == model.CUP && f.GramsPerCup == nil {
			return &model.HTTPResponse{
				Status:  http.StatusBadRequest,
				Message: "grams per cup is required for unit cup",
			}
		}

		if f.GramsPerCup != nil {
			gramsPerCupByFoodID[f.FoodID] = *f.GramsPerCup
		}

		foodIDs = append(foodIDs, f.FoodID)
	}

	petDetail, foods, resp := s.getInfoForPetFoodPlan(ctx, userID, payload.PetID, foodIDs)
	if resp != nil {
		return resp
	}

	foodsDetail := map[uint]model.Food{}
	for _, food := range foods {
		foodsDetail[food.ID] = food
	}

	foodPlanDetails := make([]*model.PetFoodPlanDetail, 0, len(payload.Foods))
	for _, food := range payload.Foods {
		foodDetail := foodsDetail[food.FoodID]
		energy := s.calculationService.CalEnergyIntakeFromGramIntake(food.Amount, foodDetail.Energy, *foodDetail.Type)
		protein, fat := s.calculationService.CalNutritientIntakeFromGramIntake(food.Amount, foodDetail.Protein, foodDetail.Fat, *foodDetail.Type)
		foodPlanDetails = append(foodPlanDetails, &model.PetFoodPlanDetail{
			Amount:        food.Amount,
			EnergyIntake:  energy,
			ProteinIntake: protein,
			FatIntake:     fat,
		})
	}

	plan := &model.PetFoodPlan{
		PetID:      payload.PetID,
		Name:       payload.Name,
		Active:     true,
		Unit:       *payload.Unit,
		SelfDefine: *payload.SelfDefine,
		CreatedAt:  time.Now(),
	}

	foodsInPlan := make([]*model.FoodPetFoodPlan, 0, len(foods))
	for _, food := range payload.Foods {
		gramsPerCup := 0.0
		if *payload.Unit == model.CUP {
			gramsPerCup = gramsPerCupByFoodID[food.FoodID]
		}

		foodsInPlan = append(foodsInPlan, &model.FoodPetFoodPlan{
			FoodID:      food.FoodID,
			GramsPerCup: gramsPerCup,
		})
	}

	foodPlanTotal := s.calculationService.CalTotalIntakeFoodPlan(foodPlanDetails)
	foodPlanTotal.PetDetailID = petDetail.ID

	if err := s.petFoodPlanRepo.CreatePetFoodPlan(ctx, payload.PetID, plan, foodsInPlan, foodPlanTotal, foodPlanDetails); err != nil {
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToCreateMsg + "pet food plan",
		}
	}

	if tier != nil && *tier == model.FREE {
		freeUsage.FoodPlanFree += 1
		s.freeValidationService.UpdateFreeTierUsage(ctx, userID, freeUsage)
	}

	activePlanDetail, err := s.petFoodPlanRepo.GetLastestActivePlanDetailByPet(ctx, payload.PetID, petDetail.ID)
	if err != nil {
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: "successfully created plan but " + utils.FailedToGetMsg + "lastest pet food plan",
		}
	}

	activePlanDetail.PetFoodPlanTotals[0].PetDetail = petDetail
	if plan.Unit == model.CUP {
		s.calculationService.ConvertGramsToCupInPlan(activePlanDetail)
	}
	return &model.HTTPResponse{
		Status: http.StatusCreated,
		Data: map[string]interface{}{
			"pet_food_plan": activePlanDetail,
		},
	}
}

func (s *petFoodPlanService) GetLastestActivePlanDetailByPet(ctx context.Context, petID uint) *model.HTTPResponse {
	petDetail, err := s.petRepo.GetLatestPetDetailByPetID(ctx, petID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &model.HTTPResponse{
				Status:  http.StatusNotFound,
				Message: "pet detail" + utils.NotFoundMsg,
			}
		}

		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToGetMsg + "pet detail",
		}
	}

	activePlanDetail, err := s.petFoodPlanRepo.GetLastestActivePlanDetailByPet(ctx, petID, petDetail.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &model.HTTPResponse{
				Status:  http.StatusNotFound,
				Message: "latest active pet food plan" + utils.NotFoundMsg,
			}
		}

		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToGetMsg + "latest active pet food plan detail",
		}
	}

	if activePlanDetail.Unit == model.CUP {
		s.calculationService.ConvertGramsToCupInPlan(activePlanDetail)
	}
	return &model.HTTPResponse{
		Status: http.StatusOK,
		Data: map[string]interface{}{
			"pet_food_plan": activePlanDetail,
		},
	}
}

func (s *petFoodPlanService) UpdateFeedingAmountFromUser(ctx context.Context, payload *model.AdjustAmountFoodInPetFoodPlanPayload) *model.HTTPResponse {
	petFoodPlan, err := s.petFoodPlanRepo.GetFoodsInLastestActivePlanByPlanID(ctx, payload.PetFoodPlanID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &model.HTTPResponse{
				Status:  http.StatusNotFound,
				Message: "latest active pet food plan" + utils.NotFoundMsg,
			}
		}

		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToGetMsg + "latest active pet food plan",
		}
	}

	if len(petFoodPlan.FoodPetFoodPlans) != len(payload.PetFoodPlanDetails) {
		return &model.HTTPResponse{
			Status:  http.StatusBadRequest,
			Message: "invalid number of food in this food plan",
		}
	}

	petDetail, err := s.petRepo.GetLatestPetDetailByPetID(ctx, petFoodPlan.PetID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &model.HTTPResponse{
				Status:  http.StatusNotFound,
				Message: "lastest pet detail" + utils.NotFoundMsg,
			}
		}

		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToGetMsg + "lastest pet detail",
		}
	}

	foodsInPlan := make(map[uint]*model.Food)
	for _, f := range petFoodPlan.FoodPetFoodPlans {
		foodsInPlan[f.ID] = f.Food
	}

	seenFoodPetFoodPlanIDs := make(map[uint]bool, len(payload.PetFoodPlanDetails))
	var foodPlanDetails []*model.PetFoodPlanDetail
	for _, fp := range payload.PetFoodPlanDetails {
		if seenFoodPetFoodPlanIDs[fp.FoodPetFoodPlanID] {
			return &model.HTTPResponse{
				Status:  http.StatusBadRequest,
				Message: "duplicate food_pet_food_plan_id is not allowed",
			}
		}
		seenFoodPetFoodPlanIDs[fp.FoodPetFoodPlanID] = true

		food, exists := foodsInPlan[fp.FoodPetFoodPlanID]
		if !exists || food == nil {
			return &model.HTTPResponse{
				Status:  http.StatusBadRequest,
				Message: "food_pet_food_plan_id is not in this pet food plan",
			}
		}

		energyInake := s.calculationService.CalEnergyIntakeFromGramIntake(fp.Amount, food.Energy, *food.Type)
		proteinIntake, fatIntake := s.calculationService.CalNutritientIntakeFromGramIntake(fp.Amount, food.Protein, food.Fat, *food.Type)
		foodPlanDetails = append(foodPlanDetails, &model.PetFoodPlanDetail{
			FoodPetFoodPlanID: fp.FoodPetFoodPlanID,
			Amount:            fp.Amount,
			EnergyIntake:      energyInake,
			ProteinIntake:     proteinIntake,
			FatIntake:         fatIntake,
		})
	}

	foodPlanTotal := s.calculationService.CalTotalIntakeFoodPlan(foodPlanDetails)
	foodPlanTotal.PetFoodPlanID = petFoodPlan.ID
	foodPlanTotal.PetDetailID = petDetail.ID

	err = s.petFoodPlanRepo.UpdateFeedingAmountFromUser(ctx, petFoodPlan.PetID, foodPlanTotal, foodPlanDetails)
	if err != nil {
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToUpdateMsg + "pet food plan",
		}
	}

	activePlanDetail, err := s.petFoodPlanRepo.GetLastestActivePlanDetailByPet(ctx, petFoodPlan.PetID, petDetail.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &model.HTTPResponse{
				Status:  http.StatusNotFound,
				Message: "latest active pet food plan" + utils.NotFoundMsg,
			}
		}

		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToGetMsg + "latest active pet food plan detail",
		}
	}

	return &model.HTTPResponse{
		Status: http.StatusOK,
		Data: map[string]interface{}{
			"pet_food_plan": activePlanDetail,
		},
	}
}
