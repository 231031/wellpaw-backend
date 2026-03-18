package repository

import (
	"context"
	"fmt"

	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/231031/wellpaw-backend/internal/utils"
	"gorm.io/gorm"
)

type FoodRepository interface {
	CreateFood(ctx context.Context, food *model.Food, foodQuantity *model.FoodQuantity) error
	GetFoodsByUserID(ctx context.Context, userID uint) ([]model.Food, error)
	GetFoodsByFoodType(ctx context.Context, userID uint, foodType model.FoodType) ([]model.Food, error)
	CreateNewFoodQuantity(ctx context.Context, foodQuantity *model.FoodQuantity) error
	UpdateFoodDetail(ctx context.Context, userID uint, foodID uint, updates map[string]interface{}) error
	GetFoodsByIDsAndUserID(ctx context.Context, userID uint, foodIDs []uint) ([]model.Food, error)
	GetFoodByIDAndUserID(ctx context.Context, userID uint, foodID uint) (*model.Food, error)
	SoftDeleteFoodByIDAndUserID(ctx context.Context, foodID uint, userID uint) error

	UpdateDailyFoodAmount(ctx context.Context, updatedAmount []model.FoodQuantity) error
}

type foodRepository struct {
	db *gorm.DB
}

func NewFoodRepository(db *gorm.DB) FoodRepository {
	return &foodRepository{
		db: db,
	}
}

func (r *foodRepository) CreateFood(ctx context.Context, food *model.Food, foodQuantity *model.FoodQuantity) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(food).Error; err != nil {
			return err
		}

		foodQuantity.FoodID = food.ID
		if err := tx.Create(foodQuantity).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *foodRepository) GetFoodsByUserID(ctx context.Context, userID uint) ([]model.Food, error) {
	var foods []model.Food
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Preload("FoodQuantities", func(db *gorm.DB) *gorm.DB {
			return db.Where("amount > ?", 0).Order("created_at ASC")
		}).
		Order("brand DESC, id DESC").
		Find(&foods).Error; err != nil {
		return nil, fmt.Errorf("failed to get foods by user id : %w", err)
	}

	return foods, nil
}

func (r *foodRepository) GetFoodsByFoodType(ctx context.Context, userID uint, foodType model.FoodType) ([]model.Food, error) {
	var foods []model.Food
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND type = ?", userID, foodType).
		Preload("FoodQuantities", func(db *gorm.DB) *gorm.DB {
			return db.Where("amount > ?", 0).Order("created_at ASC")
		}).
		Order("brand DESC, id DESC").
		Find(&foods).Error; err != nil {
		return nil, fmt.Errorf("failed to get foods by food type : %w", err)
	}

	return foods, nil
}

func (r *foodRepository) CreateNewFoodQuantity(ctx context.Context, foodQuantity *model.FoodQuantity) error {
	result := r.db.WithContext(ctx).Create(foodQuantity)
	if result.Error != nil {
		return result.Error
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

	foodsByID := make(map[uint]model.Food, len(foods))
	for _, f := range foods {
		foodsByID[f.ID] = f
	}

	ordered := make([]model.Food, 0, len(foodIDs))
	for _, id := range foodIDs {
		if f, ok := foodsByID[id]; ok {
			ordered = append(ordered, f)
		}
	}
	foods = ordered

	return foods, nil
}

func (r *foodRepository) GetFoodByIDAndUserID(ctx context.Context, userID uint, foodID uint) (*model.Food, error) {
	var food *model.Food

	if err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", foodID, userID).
		Preload("FoodQuantities", func(db *gorm.DB) *gorm.DB {
			return db.Where("amount > ?", 0).Order("created_at ASC")
		}).
		Find(&food).Error; err != nil {
		return nil, fmt.Errorf("failed to get food by ids and user id : %w", err)
	}

	return food, nil
}

func (r *foodRepository) UpdateFoodDetail(ctx context.Context, userID uint, foodID uint, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return utils.ErrNoRowsUpdated
	}

	result := r.db.WithContext(ctx).
		Model(&model.Food{}).
		Where("id = ? AND user_id = ?", foodID, userID).
		Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update food detail : %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return utils.ErrNoRowsUpdated
	}

	return nil
}

func (r *foodRepository) UpdateDailyFoodAmount(ctx context.Context, updatedAmount []model.FoodQuantity) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, q := range updatedAmount {
			err := tx.Model(&model.FoodQuantity{}).Where("id = ?", q.ID).
				Updates(map[string]interface{}{"amount": q.Amount}).Error
			if err != nil {
				return err
			}
		}

		return nil
	})
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
