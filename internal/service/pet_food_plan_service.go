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
	CreatePetFoodPlan(ctx context.Context, userID uint, payload *model.CreatePetFoodPlanPayload) *model.HTTPResponse
	GetLastestActivePlanDetailByPet(ctx context.Context, petID uint) *model.HTTPResponse
	UpdateFeedingAmountFromUser(ctx context.Context, payload *model.AdjustAmountFoodInPetFoodPlanPayload) *model.HTTPResponse
}

type petFoodPlanService struct {
	calculationService CalculationService
	petFoodPlanRepo    repository.PetFoodPlanRepository
	petRepo            repository.PetRepository
	foodRepo           repository.FoodRepository
}

func NewPetFoodPlanService(calculationService CalculationService, petFoodPlanRepo repository.PetFoodPlanRepository, petRepo repository.PetRepository, foodRepo repository.FoodRepository) PetFoodPlanService {
	return &petFoodPlanService{
		calculationService: calculationService,
		petFoodPlanRepo:    petFoodPlanRepo,
		petRepo:            petRepo,
		foodRepo:           foodRepo,
	}
}

func (s *petFoodPlanService) CreatePetFoodPlan(ctx context.Context, userID uint, payload *model.CreatePetFoodPlanPayload) *model.HTTPResponse {
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

	petDetail, err := s.petRepo.GetLatestPetDetailByPetID(ctx, payload.PetID)
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

	foods, err := s.foodRepo.GetFoodsByIDsAndUserID(ctx, userID, foodIDs)
	if err != nil {
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToGetMsg + "foods",
		}
	}

	if len(foods) != len(foodIDs) {
		return &model.HTTPResponse{
			Status:  http.StatusNotFound,
			Message: "some foods" + utils.NotFoundMsg,
		}
	}

	foodPlanDetails := s.calculationService.CalFeedingAmountPerDay(petDetail, foods)
	if len(foodPlanDetails) != len(foods) {
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: "failed to calculate feeding plan details",
		}
	}

	plan := &model.PetFoodPlan{
		PetID:     payload.PetID,
		Name:      payload.Name,
		Active:    true,
		Unit:      *payload.Unit,
		CreatedAt: time.Now(),
	}

	foodsInPlan := make([]*model.FoodPetFoodPlan, 0, len(foods))
	cupFoods := make([]*model.CupFoodPet, 0, len(foods))
	for _, food := range foods {
		foodsInPlan = append(foodsInPlan, &model.FoodPetFoodPlan{
			FoodID: food.ID,
		})

		if *payload.Unit == model.CUP {
			cupFoods = append(cupFoods, &model.CupFoodPet{
				Grams: gramsPerCupByFoodID[food.ID],
			})
		}
	}

	foodPlanTotal := s.calculationService.CalTotalIntakeFoodPlan(foodPlanDetails)
	foodPlanTotal.PetDetailID = petDetail.ID

	if err := s.petFoodPlanRepo.CreatePetFoodPlan(ctx, payload.PetID, plan, foodsInPlan, foodPlanTotal, foodPlanDetails, cupFoods); err != nil {
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToCreateMsg + "pet food plan",
		}
	}

	activePlanDetail, err := s.petFoodPlanRepo.GetLastestActivePlanDetailByPet(ctx, payload.PetID, petDetail.ID)
	if err != nil {
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: "successfully created plan but " + utils.FailedToGetMsg + "lastest pet food plan",
		}
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
