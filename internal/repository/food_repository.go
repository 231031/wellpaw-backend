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
	SoftDeleteFoodByIDAndUserID(ctx context.Context, foodID uint, userID uint) error
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

func (r *foodRepository) SoftDeleteFoodByIDAndUserID(ctx context.Context, foodID uint, userID uint) error {
	var activePlanCount int64
	if err := r.db.WithContext(ctx).
		Table("food_pet_food_plans").
		Joins("JOIN pet_food_plans ON pet_food_plans.id = food_pet_food_plans.pet_food_plan_id").
		Joins("JOIN foods ON foods.id = food_pet_food_plans.food_id").
		Where("food_pet_food_plans.food_id = ? AND foods.user_id = ? AND foods.deleted_at IS NULL AND pet_food_plans.active = ?", foodID, userID, true).
		Count(&activePlanCount).Error; err != nil {
		return fmt.Errorf("failed to check active plan before deleting food : %w", err)
	}

	if activePlanCount > 0 {
		return utils.ErrFoodInActivePlan
	}

	result := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", foodID, userID).
		Delete(&model.Food{})
	if result.Error != nil {
		return fmt.Errorf("failed to soft delete food : %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return utils.ErrNoRowsUpdated
	}

	return nil
}
