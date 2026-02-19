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

func (s *foodService) UpdateFoodWeightAndQuantity(ctx context.Context, userID uint, id uint, weight float64, quantity int) error {
	// logic to update food weight and quantity
	return s.foodRepo.UpdateFoodWeightAndQuantity(ctx, userID, id, weight, quantity)
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
