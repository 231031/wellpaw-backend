package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/231031/wellpaw-backend/internal/model"
	"gorm.io/gorm"
)

type PetRepository interface {
	CreateNewPet(ctx context.Context, pet *model.Pet, petDetails *model.PetDetail) error
	UpdatePetDetails(ctx context.Context, petDetails *model.PetDetail) error
}

type petRepository struct {
	db *gorm.DB
}

func NewPetRepository(db *gorm.DB) PetRepository {
	return &petRepository{
		db: db,
	}
}

func (r *petRepository) CreateNewPet(ctx context.Context, pet *model.Pet, petDetails *model.PetDetail) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(pet).Error; err != nil {
			return fmt.Errorf("failed to create new pet : %w", err)
		}

		petDetails.PetID = pet.ID
		if err := tx.Create(petDetails).Error; err != nil {
			return fmt.Errorf("failed to create pet details : %w", err)
		}
		return nil
	})

	if err != nil {
		return err
	}
	return nil
}

// increase new row to collect all pet details history
func (r *petRepository) UpdatePetDetails(ctx context.Context, petDetails *model.PetDetail) error {
	if err := r.db.WithContext(ctx).Create(petDetails).Error; err != nil {
		return fmt.Errorf("failed to create pet details : %w", err)
	}

	return nil
}

func (r *petRepository) GetPetByID(ctx context.Context, id uint) (*model.Pet, error) {
	currentDate := time.Now()
	oneYearAgo := currentDate.AddDate(-1, 0, 0)

	// use Time-Weighted Average - TWA to cal avg per month
	var pet *model.Pet
	err := r.db.WithContext(ctx).Preload("PetDetails", func(db *gorm.DB) *gorm.DB {
		return db.Preload("PetFoodPlanDetails.FoodPetFoodPlans.PetFoodPlans").Where("created_at >= ?", oneYearAgo).Order("created_at asc")
	}).First(&pet, id).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get pet by id : %w", err)
	}

	return pet, nil
}

func (r *petRepository) GetPetByUserID(ctx context.Context, id uint) ([]model.Pet, error) {
	var pets []model.Pet
	err := r.db.WithContext(ctx).Where("user_id = ?", id).Find(&pets).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get pet by user id : %w", err)
	}

	return pets, nil
}
