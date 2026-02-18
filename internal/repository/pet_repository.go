package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/231031/wellpaw-backend/internal/utils"
	"gorm.io/gorm"
)

type PetRepository interface {
	CreateNewPet(ctx context.Context, pet *model.Pet, petDetails *model.PetDetail) error
	UpdatePetInfo(ctx context.Context, pet *model.Pet) error
	UpdatePetDetails(ctx context.Context, petDetails *model.PetDetail) error
	UpdatePetDetailsAndPlan(ctx context.Context, petID uint, petDetails *model.PetDetail, foodPlanTotal *model.PetFoodPlanTotal, foodPlanDetails []*model.PetFoodPlanDetail) error
	GetPetInfoByID(ctx context.Context, id uint) (*model.Pet, error)
	GetLatestPetDetailByPetID(ctx context.Context, petID uint) (*model.PetDetail, error)
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

func (r *petRepository) UpdatePetInfo(ctx context.Context, pet *model.Pet) error {
	result := r.db.WithContext(ctx).Model(&model.Pet{}).
		Where("id = ?", pet.ID).
		Updates(map[string]interface{}{
			"image_path": pet.ImagePath,
			"name":       pet.Name,
		})
	if result.Error != nil {
		return fmt.Errorf("failed to update pet info : %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return utils.ErrNoRowsUpdated
	}

	return nil
}

func (r *petRepository) UpdatePetDetails(ctx context.Context, petDetails *model.PetDetail) error {
	if err := r.db.Create(petDetails).Error; err != nil {
		return err
	}

	return nil
}

func (r *petRepository) UpdatePetDetailsAndPlan(ctx context.Context, petID uint, petDetails *model.PetDetail, foodPlanTotal *model.PetFoodPlanTotal, foodPlanDetails []*model.PetFoodPlanDetail) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(petDetails).Error; err != nil {
			return fmt.Errorf("failed to update pet details : %w", err)
		}

		foodPlanTotal.PetDetailID = petDetails.ID
		if err := tx.Create(foodPlanTotal).Error; err != nil {
			return err
		}

		for idx := range foodPlanDetails {
			foodPlanDetails[idx].PetFoodPlanTotalID = foodPlanTotal.ID
		}
		if err := tx.Create(foodPlanDetails).Error; err != nil {
			return fmt.Errorf("failed to update food plan detail : %w", err)
		}

		planHistory := &model.PetFoodPlanHistory{
			PetID:              petID,
			PetFoodPlanTotalID: foodPlanTotal.ID,
		}
		if err := tx.Create(planHistory).Error; err != nil {
			return fmt.Errorf("failed to update new history of food : %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *petRepository) GetPetInfoByID(ctx context.Context, id uint) (*model.Pet, error) {
	var pet *model.Pet
	err := r.db.WithContext(ctx).First(&pet, id).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get pet by id : %w", err)
	}

	return pet, nil
}

func (r *petRepository) GetLatestPetDetailByPetID(ctx context.Context, petID uint) (*model.PetDetail, error) {
	var petDetail model.PetDetail
	if err := r.db.WithContext(ctx).
		Where("pet_id = ?", petID).
		Order("created_at DESC, id DESC").
		First(&petDetail).Error; err != nil {
		return nil, fmt.Errorf("failed to get latest pet detail by pet id : %w", err)
	}

	return &petDetail, nil
}

func (r *petRepository) GetPetByID(ctx context.Context, id uint) (*model.Pet, error) {
	currentDate := time.Now()
	oneYearAgo := currentDate.AddDate(-1, 0, 0)

	// use Time-Weighted Average - TWA to cal avg per month
	var pet *model.Pet
	err := r.db.WithContext(ctx).Preload("PetDetails", func(db *gorm.DB) *gorm.DB {
		return db.Preload("PetFoodPlanTotals.PetFoodPlanDetails.FoodPetFoodPlan.Food").Where("created_at >= ?", oneYearAgo).Order("created_at asc")
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
