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
	UpdateFoodWeightAndQuantity(ctx context.Context, userID uint, id uint, weight float64, quantity int) error
	SoftDeleteFood(ctx context.Context, userID uint, foodID uint) *model.HTTPResponse
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
	food, err := s.foodRepo.CreateFood(ctx, food)
	if err != nil {
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToCreateMsg + "food",
		}
	}

	return &model.HTTPResponse{
		Status: http.StatusCreated,
		Data:   food,
	}
}

func (s *foodService) GetFoodsByUserID(ctx context.Context, userID uint) *model.HTTPResponse {
	foods, err := s.foodRepo.GetFoodsByUserID(ctx, userID)
	if err != nil {
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToGetMsg + "foods",
		}
	}

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

	return &model.HTTPResponse{
		Status: http.StatusOK,
		Data: map[string]interface{}{
			"foods": foods,
		},
	}
}

func (s *foodService) UpdateFoodWeightAndQuantity(ctx context.Context, userID uint, id uint, weight float64, quantity int) error {
	// logic to update food weight and quantity
	return s.foodRepo.UpdateFoodWeightAndQuantity(ctx, userID, id, weight, quantity)
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
