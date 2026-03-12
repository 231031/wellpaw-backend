package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/231031/wellpaw-backend/internal/model"
	"gorm.io/gorm"
)

type PetFoodPlanRepository interface {
	CreatePetFoodPlan(ctx context.Context, petID uint, plan *model.PetFoodPlan, foodsInFoodPlan []*model.FoodPetFoodPlan, foodPlanTotal *model.PetFoodPlanTotal, intakeDetailPerFood []*model.PetFoodPlanDetail) error
	GetFoodsInLastestActivePlanByPetID(ctx context.Context, petID uint) (uint, []model.FoodPetFoodPlan, error)
	GetFoodsInLastestActivePlanByPlanID(ctx context.Context, planID uint) (*model.PetFoodPlan, error)
	GetLastestActivePlanDetailByPet(ctx context.Context, petID uint, petDetailID uint) (*model.PetFoodPlan, error)
	GetPlanUsageHistoryByPetID(ctx context.Context, petID uint) ([]model.PetFoodPlanHistory, error)
	GetNextActiveFoodPlanByID(ctx context.Context, lastID uint) (*model.PetFoodPlan, error)
	InactivePetFoodPlan(ctx context.Context, plans []model.PetFoodPlan) error
	UpdateFeedingAmountFromUser(ctx context.Context, petID uint, foodPlanTotal *model.PetFoodPlanTotal, foodPlanDetails []*model.PetFoodPlanDetail) error
}

type petFoodPlanRepository struct {
	db *gorm.DB
}

func NewPetFoodPlanRepository(db *gorm.DB) PetFoodPlanRepository {
	return &petFoodPlanRepository{
		db: db,
	}
}

func (r *petFoodPlanRepository) CreatePetFoodPlan(ctx context.Context, petID uint, plan *model.PetFoodPlan, foodsInFoodPlan []*model.FoodPetFoodPlan, foodPlanTotal *model.PetFoodPlanTotal, foodPlanDetails []*model.PetFoodPlanDetail) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.PetFoodPlan{}).Where("pet_id = ? AND active = ?", plan.PetID, true).Update("active", false).Error; err != nil {
			return err
		}

		if err := tx.Create(plan).Error; err != nil {
			return err
		}

		for _, foodInPlan := range foodsInFoodPlan {
			foodInPlan.PetFoodPlanID = plan.ID
		}
		if err := tx.Create(foodsInFoodPlan).Error; err != nil {
			return err
		}

		if len(foodPlanDetails) != len(foodsInFoodPlan) {
			return fmt.Errorf("invalid plan detail size: details=%d foods=%d", len(foodPlanDetails), len(foodsInFoodPlan))
		}

		foodPlanTotal.PetFoodPlanID = plan.ID
		if err := tx.Create(foodPlanTotal).Error; err != nil {
			return err
		}

		for idx, detail := range foodPlanDetails {
			detail.FoodPetFoodPlanID = foodsInFoodPlan[idx].ID
			detail.PetFoodPlanTotalID = foodPlanTotal.ID
		}
		if err := tx.Create(foodPlanDetails).Error; err != nil {
			return err
		}

		foodPlanHistory := &model.PetFoodPlanHistory{
			PetID:              petID,
			PetFoodPlanTotalID: foodPlanTotal.ID,
		}
		if err := tx.Create(foodPlanHistory).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *petFoodPlanRepository) GetFoodsInLastestActivePlanByPetID(ctx context.Context, petID uint) (uint, []model.FoodPetFoodPlan, error) {
	var latestPlanID uint
	err := r.db.WithContext(ctx).
		Model(&model.PetFoodPlan{}).
		Select("pet_food_plans.id").
		Where("pet_id = ? AND active = ?", petID, true).
		Order("created_at DESC, id DESC").
		Limit(1).
		Scan(&latestPlanID).Error
	if err != nil {
		return 0, nil, err
	}
	if latestPlanID == 0 {
		return 0, nil, gorm.ErrRecordNotFound
	}

	var foodsInPlan []model.FoodPetFoodPlan
	if err := r.db.WithContext(ctx).
		Model(&model.FoodPetFoodPlan{}).
		Preload("Food").
		Where("pet_food_plan_id = ?", latestPlanID).
		Order("food_pet_food_plans.id ASC").
		Find(&foodsInPlan).Error; err != nil {
		return 0, nil, fmt.Errorf("failed to get foods from latest active pet food plan : %w", err)
	}

	if len(foodsInPlan) == 0 {
		return 0, nil, gorm.ErrRecordNotFound
	}

	return latestPlanID, foodsInPlan, nil
}

func (r *petFoodPlanRepository) GetFoodsInLastestActivePlanByPlanID(ctx context.Context, planID uint) (*model.PetFoodPlan, error) {
	var petFoodPlan model.PetFoodPlan
	if err := r.db.WithContext(ctx).
		Where("id = ? AND active =?", planID, true).
		Preload("FoodPetFoodPlans", func(db *gorm.DB) *gorm.DB {
			return db.Order("id ASC")
		}).
		Preload("FoodPetFoodPlans.Food").
		First(&petFoodPlan).Error; err != nil {
		return nil, fmt.Errorf("failed to get foods from latest active pet food plan : %w", err)
	}

	return &petFoodPlan, nil
}

func (r *petFoodPlanRepository) GetLastestActivePlanDetailByPet(ctx context.Context, petID uint, petDetailID uint) (*model.PetFoodPlan, error) {
	var foodPlan model.PetFoodPlan
	if err := r.db.WithContext(ctx).
		Where("pet_id = ? AND active = ?", petID, true).
		Order("created_at DESC, id DESC").
		Preload("PetFoodPlanTotals", func(db *gorm.DB) *gorm.DB {
			return db.Where("pet_detail_id = ?", petDetailID).
				Order("created_at DESC, id DESC").
				Limit(1)
		}).
		Preload("PetFoodPlanTotals.PetFoodPlanDetails", func(db *gorm.DB) *gorm.DB {
			return db.Order("id ASC")
		}).
		Preload("PetFoodPlanTotals.PetFoodPlanDetails.FoodPetFoodPlan").
		Preload("PetFoodPlanTotals.PetFoodPlanDetails.FoodPetFoodPlan.Food").
		First(&foodPlan).Error; err != nil {
		return nil, fmt.Errorf("failed to get latest active pet food plan detail : %w", err)
	}

	return &foodPlan, nil
}

func (r *petFoodPlanRepository) GetPlanUsageHistoryByPetID(ctx context.Context, petID uint) ([]model.PetFoodPlanHistory, error) {
	var histories []model.PetFoodPlanHistory
	if err := r.db.WithContext(ctx).
		Preload("PetFoodPlanTotal").
		Preload("PetFoodPlanTotal.PetFoodPlan").
		Preload("PetFoodPlanTotal.PetFoodPlanDetails", func(db *gorm.DB) *gorm.DB {
			return db.Order("id ASC")
		}).
		Preload("PetFoodPlanTotal.PetFoodPlanDetails.FoodPetFoodPlan").
		Preload("PetFoodPlanTotal.PetFoodPlanDetails.FoodPetFoodPlan.Food").
		Where("pet_id = ?", petID).
		Order("created_at DESC, id DESC").
		Limit(2).
		Find(&histories).Error; err != nil {
		return nil, fmt.Errorf("failed to get plan usage history by pet id : %w", err)
	}

	if len(histories) == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	now := time.Now()
	for idx := range histories {
		if idx == 0 {
			histories[idx].PlanUsageEndDate = now
			continue
		}
		histories[idx].PlanUsageEndDate = histories[idx-1].CreatedAt
	}

	return histories, nil
}

// not test
func (r *petFoodPlanRepository) UpdateFeedingAmountFromUser(ctx context.Context, petID uint, foodPlanTotal *model.PetFoodPlanTotal, foodPlanDetails []*model.PetFoodPlanDetail) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(foodPlanTotal).Error; err != nil {
			return fmt.Errorf("failed to create pet food plan total : %w", err)
		}

		for idx, detail := range foodPlanDetails {
			if detail == nil {
				return fmt.Errorf("food plan detail at index %d is nil", idx)
			}
			detail.PetFoodPlanTotalID = foodPlanTotal.ID
		}
		if err := tx.Create(foodPlanDetails).Error; err != nil {
			return fmt.Errorf("failed to create pet food plan details : %w", err)
		}

		foodPlanHistory := &model.PetFoodPlanHistory{
			PetID:              petID,
			PetFoodPlanTotalID: foodPlanTotal.ID,
		}
		if err := tx.Create(foodPlanHistory).Error; err != nil {
			return fmt.Errorf("failed to create pet food plan history : %w", err)
		}

		return nil
	})
}

func (r *petFoodPlanRepository) GetNextActiveFoodPlanByID(ctx context.Context, lastID uint) (*model.PetFoodPlan, error) {
	var foodPlan model.PetFoodPlan

	err := r.db.WithContext(ctx).Model(&model.PetFoodPlan{}).
		Where("active = true AND id > ?", lastID).
		Order("id ASC").
		Limit(1).
		Preload("Pet").
		Preload("Pet.User").
		Preload("FoodPetFoodPlans", func(db *gorm.DB) *gorm.DB {
			return db.Order("id ASC")
		}).
		Preload("FoodPetFoodPlans.Food").
		Preload("FoodPetFoodPlans.Food.FoodQuantities", func(db *gorm.DB) *gorm.DB {
			return db.Where("amount >= 0").Order("created_at ASC")
		}).
		Preload("FoodPetFoodPlans.PetFoodPlanDetails", func(db *gorm.DB) *gorm.DB {
			return db.Where("id IN (SELECT MAX(id) FROM pet_food_plan_details GROUP BY food_pet_food_plan_id)")
		}).
		First(&foodPlan).Error
	if err != nil {
		return nil, err
	}

	return &foodPlan, nil
}

func (r *petFoodPlanRepository) InactivePetFoodPlan(ctx context.Context, plans []model.PetFoodPlan) error {
	ids := make([]uint, 0, len(plans))
	for _, p := range plans {
		ids = append(ids, p.ID)
	}

	return r.db.WithContext(ctx).
		Model(&model.PetFoodPlan{}).
		Where("id IN ?", ids).
		Update("active", false).Error
}

// func (r *petFoodPlanRepository) UpdateActivePetFoodPlan(ctx context.Context, petFoodPlanID uint) error {
// 	return nil
// }
