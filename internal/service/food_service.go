package service

import (
	"context"
	"net/http"

	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/231031/wellpaw-backend/internal/repository"
	"github.com/231031/wellpaw-backend/internal/utils"
)

type FoodService interface {
	CreateFood(ctx context.Context, food *model.Food) *model.HTTPResponse
	UpdateFoodWeightAndQuantity(ctx context.Context, id uint, weight float64, quantity int) error
}

type foodService struct {
	foodRepo repository.FoodRepository
}

func NewFoodService(foodRepo repository.FoodRepository) FoodService {
	return &foodService{
		foodRepo: foodRepo,
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

func (s *foodService) UpdateFoodWeightAndQuantity(ctx context.Context, id uint, weight float64, quantity int) error {
	// logic to update food weight and quantity
	return s.foodRepo.UpdateFoodWeightAndQuantity(ctx, id, weight, quantity)
}
