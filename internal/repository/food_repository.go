package repository

import (
	"context"

	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/231031/wellpaw-backend/internal/utils"
	"gorm.io/gorm"
)

type FoodRepository interface {
	CreateFood(ctx context.Context, food *model.Food) (*model.Food, error)
	UpdateFoodWeightAndQuantity(ctx context.Context, id uint, weight float64, quantity int) error
}

type foodRepository struct {
	db *gorm.DB
}

func NewFoodRepository(db *gorm.DB) FoodRepository {
	return &foodRepository{
		db: db,
	}
}

func (r *foodRepository) CreateFood(ctx context.Context, food *model.Food) (*model.Food, error) {
	if err := r.db.WithContext(ctx).Create(food).Error; err != nil {
		return nil, err
	}
	return food, nil
}

func (r *foodRepository) UpdateFoodWeightAndQuantity(ctx context.Context, id uint, weight float64, quantity int) error {
	food := model.Food{
		ID:       id,
		Weight:   weight,
		Quantity: quantity,
	}

	result := r.db.WithContext(ctx).Updates(food)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return utils.ErrNoRowsUpdated
	}

	return nil
}
