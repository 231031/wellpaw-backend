package service

import (
	"context"
	"errors"
	"net/http"

	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/231031/wellpaw-backend/internal/repository"
	"github.com/231031/wellpaw-backend/internal/utils"
)

type FoodService interface {
	CreateFood(ctx context.Context, food *model.Food) *model.HTTPResponse
	GetFoodsByUserID(ctx context.Context, userID uint) *model.HTTPResponse
	GetFoodsByFoodType(ctx context.Context, userID uint, foodType model.FoodType) *model.HTTPResponse
	UpdateFoodDetail(ctx context.Context, userID uint, foodID uint, payload *model.UpdateFoodDetailPayload) *model.HTTPResponse
	CreateNewFoodQuantity(ctx context.Context, userID uint, payload *model.FoodQuantity) *model.HTTPResponse
	SoftDeleteFood(ctx context.Context, userID uint, foodID uint) *model.HTTPResponse

	mapCurrentFoodQuantity(foods []model.Food) []model.Food
}

type foodService struct {
	calculationService CalculationService
	foodRepo           repository.FoodRepository
}

func NewFoodService(calculationService CalculationService, foodRepo repository.FoodRepository) FoodService {
	return &foodService{
		calculationService: calculationService,
		foodRepo:           foodRepo,
	}
}

func (s *foodService) CreateFood(ctx context.Context, food *model.Food) *model.HTTPResponse {
	foodQuantity := &model.FoodQuantity{
		Weight:   food.Weight,
		Quantity: food.Quantity,
		Amount:   float64(food.Quantity) * food.Weight,
	}

	err := s.foodRepo.CreateFood(ctx, food, foodQuantity)
	if err != nil {
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToCreateMsg + "food",
		}
	}

	food.FoodQuantities = append(food.FoodQuantities, *foodQuantity)
	return &model.HTTPResponse{
		Status: http.StatusCreated,
		Data:   food,
	}
}

func (s *foodService) mapCurrentFoodQuantity(foods []model.Food) []model.Food {
	for idx := range foods {
		if len(foods[idx].FoodQuantities) > 0 {
			foods[idx].Weight = foods[idx].FoodQuantities[0].Weight
			foods[idx].Quantity = foods[idx].FoodQuantities[0].Quantity
		}
	}
	return foods
}

func (s *foodService) GetFoodsByUserID(ctx context.Context, userID uint) *model.HTTPResponse {
	foods, err := s.foodRepo.GetFoodsByUserID(ctx, userID)
	if err != nil {
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToGetMsg + "foods",
		}
	}

	foods = s.mapCurrentFoodQuantity(foods)
	return &model.HTTPResponse{
		Status: http.StatusOK,
		Data: map[string]interface{}{
			"foods": foods,
		},
	}
}

func (s *foodService) GetFoodsByFoodType(ctx context.Context, userID uint, foodType model.FoodType) *model.HTTPResponse {
	foods, err := s.foodRepo.GetFoodsByFoodType(ctx, userID, foodType)
	if err != nil {
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToGetMsg + "foods",
		}
	}

	foods = s.mapCurrentFoodQuantity(foods)
	return &model.HTTPResponse{
		Status: http.StatusOK,
		Data: map[string]interface{}{
			"foods": foods,
		},
	}
}

func (s *foodService) CreateNewFoodQuantity(ctx context.Context, userID uint, payload *model.FoodQuantity) *model.HTTPResponse {
	payload.Amount = payload.Weight * float64(payload.Quantity)
	if err := s.foodRepo.CreateNewFoodQuantity(ctx, payload); err != nil {
		if errors.Is(err, utils.ErrNoRowsUpdated) {
			return &model.HTTPResponse{
				Status:  http.StatusNotFound,
				Message: "food" + utils.NotFoundMsg,
			}
		}

		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToUpdateMsg + "food detail",
		}
	}

	food, err := s.foodRepo.GetFoodByIDAndUserID(ctx, userID, payload.FoodID)
	if err != nil {
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: "food add quantity, " + utils.FailedToGetMsg + "food detail",
		}
	}
	if len(food.FoodQuantities) > 0 {
		food.Weight = food.FoodQuantities[0].Weight
		food.Quantity = food.FoodQuantities[0].Quantity
	}

	return &model.HTTPResponse{
		Status: http.StatusCreated,
		Data: map[string]interface{}{
			"food": food,
		},
	}
}

func (s *foodService) UpdateFoodDetail(ctx context.Context, userID uint, foodID uint, payload *model.UpdateFoodDetailPayload) *model.HTTPResponse {
	updates := make(map[string]interface{})

	if payload.Name != nil {
		updates["name"] = *payload.Name
	}

	if payload.ImagePath != nil {
		updates["image_path"] = *payload.ImagePath
	}
	if payload.Weight != nil {
		updates["weight"] = *payload.Weight
	}
	if payload.Quantity != nil {
		updates["quantity"] = *payload.Quantity
	}

	if len(updates) == 0 {
		return &model.HTTPResponse{
			Status:  http.StatusBadRequest,
			Message: "no information to update",
		}
	}

	if err := s.foodRepo.UpdateFoodDetail(ctx, userID, foodID, updates); err != nil {
		if errors.Is(err, utils.ErrNoRowsUpdated) {
			return &model.HTTPResponse{
				Status:  http.StatusNotFound,
				Message: "food" + utils.NotFoundMsg,
			}
		}

		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToUpdateMsg + "food detail",
		}
	}

	food, err := s.foodRepo.GetFoodByIDAndUserID(ctx, userID, foodID)
	if err != nil {
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: "food updated, " + utils.FailedToGetMsg + "food detail",
		}
	}
	if len(food.FoodQuantities) > 0 {
		food.Weight = food.FoodQuantities[0].Weight
		food.Quantity = food.FoodQuantities[0].Quantity
	}

	return &model.HTTPResponse{
		Status: http.StatusOK,
		Data: map[string]interface{}{
			"food": food,
		},
	}
}

func (s *foodService) SoftDeleteFood(ctx context.Context, userID uint, foodID uint) *model.HTTPResponse {
	if err := s.foodRepo.SoftDeleteFoodByIDAndUserID(ctx, foodID, userID); err != nil {
		if errors.Is(err, utils.ErrFoodInActivePlan) {
			return &model.HTTPResponse{
				Status:  http.StatusBadRequest,
				Message: err.Error(),
			}
		}

		if errors.Is(err, utils.ErrNoRowsUpdated) {
			return &model.HTTPResponse{
				Status:  http.StatusNotFound,
				Message: "food" + utils.NotFoundMsg,
			}
		}

		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: "failed to delete food",
		}
	}

	return &model.HTTPResponse{
		Status:  http.StatusOK,
		Message: "food deleted successfully",
	}
}
