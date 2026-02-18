package repository

import (
	"context"
	"fmt"

	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/231031/wellpaw-backend/internal/utils"
	"gorm.io/gorm"
)

type FoodRepository interface {
	CreateFood(ctx context.Context, food *model.Food) (*model.Food, error)
	UpdateFoodWeightAndQuantity(ctx context.Context, userID uint, id uint, weight float64, quantity int) error
	GetFoodsByIDsAndUserID(ctx context.Context, userID uint, foodIDs []uint) ([]model.Food, error)
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

// not test
func (r *foodRepository) UpdateFoodWeightAndQuantity(ctx context.Context, userID uint, id uint, weight float64, quantity int) error {
	result := r.db.WithContext(ctx).
		Model(&model.Food{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(map[string]interface{}{
			"weight":   weight,
			"quantity": quantity,
		})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return utils.ErrNoRowsUpdated
	}

	return nil
}

func (r *foodRepository) GetFoodsByIDsAndUserID(ctx context.Context, userID uint, foodIDs []uint) ([]model.Food, error) {
	var foods []model.Food
	if len(foodIDs) == 0 {
		return foods, nil
	}

	if err := r.db.WithContext(ctx).
		Where("id IN ? AND user_id = ?", foodIDs, userID).
		Find(&foods).Error; err != nil {
		return nil, fmt.Errorf("failed to get foods by ids and user id : %w", err)
	}

	return foods, nil
}
