package repository

import (
	"context"

	"github.com/231031/wellpaw-backend/internal/model"
	"gorm.io/gorm"
)

type PetFoodPlanRepository interface {
}

type petFoodPlanRepository struct {
	db *gorm.DB
}

func NewPetFoodPlanRepository(db *gorm.DB) PetFoodPlanRepository {
	return &petFoodPlanRepository{
		db: db,
	}
}

func (r *petFoodPlanRepository) CreatePetFoodPlanDetails(ctx context.Context, plan *model.PetFoodPlan, foods []model.FoodPetFoodPlan, intakeDetail []model.PetFoodPlanDetail) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(plan).Error; err != nil {
			return err
		}
		if err := tx.Create(foods).Error; err != nil {
			return err
		}
		if err := tx.Create(intakeDetail).Error; err != nil {
			return err
		}

		return nil
	})
}
