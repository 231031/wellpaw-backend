package repository

import (
	"context"
	"fmt"

	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/231031/wellpaw-backend/internal/utils"
	"gorm.io/gorm"
)

type PetSkinImageRepository interface {
	CreatePetSkinImage(ctx context.Context, petSkinImage *model.PetSkinImage) error
	GetPetSkinImagesByUserID(ctx context.Context, userID uint) ([]model.PetSkinImage, error)
	GetPetSkinImagesByPetIDAndUserID(ctx context.Context, petID uint, userID uint) ([]model.PetSkinImage, error)
	GetPetSkinImageByID(ctx context.Context, petSkinImageID uint) (*model.PetSkinImage, error)
	UpdateLabeledPetSkinDiseaseByID(ctx context.Context, petSkinImageID uint, petID uint, userID uint, labeled model.DiseaseType, imageEvidence string) error
}

type petSkinImageRepository struct {
	db *gorm.DB
}

func NewPetSkinImageRepository(db *gorm.DB) PetSkinImageRepository {
	return &petSkinImageRepository{
		db: db,
	}
}

func (r *petSkinImageRepository) CreatePetSkinImage(ctx context.Context, petSkinImage *model.PetSkinImage) error {
	if err := r.db.WithContext(ctx).Create(petSkinImage).Error; err != nil {
		return fmt.Errorf("failed to create pet skin image : %w", err)
	}

	return nil
}

func (r *petSkinImageRepository) GetPetSkinImagesByUserID(ctx context.Context, userID uint) ([]model.PetSkinImage, error) {
	var petSkinImages []model.PetSkinImage

	if err := r.getPetSkinImagesByUserIDQuery(ctx, userID).
		Order("pet_skin_images.created_at DESC, pet_skin_images.id DESC").
		Find(&petSkinImages).Error; err != nil {
		return nil, fmt.Errorf("failed to get pet skin images by user id : %w", err)
	}

	return petSkinImages, nil
}

func (r *petSkinImageRepository) GetPetSkinImagesByPetIDAndUserID(ctx context.Context, petID uint, userID uint) ([]model.PetSkinImage, error) {
	var petSkinImages []model.PetSkinImage

	if err := r.getPetSkinImagesByUserIDQuery(ctx, userID).
		Where("pet_skin_images.pet_id = ?", petID).
		Order("pet_skin_images.created_at DESC, pet_skin_images.id DESC").
		Find(&petSkinImages).Error; err != nil {
		return nil, fmt.Errorf("failed to get pet skin images by pet id and user id : %w", err)
	}

	return petSkinImages, nil
}

func (r *petSkinImageRepository) getPetSkinImagesByUserIDQuery(ctx context.Context, userID uint) *gorm.DB {
	return r.db.WithContext(ctx).
		Model(&model.PetSkinImage{}).
		Joins("JOIN pets ON pets.id = pet_skin_images.pet_id AND pets.deleted_at IS NULL").
		Where("pets.user_id = ?", userID).
		Preload("Pet")
}

func (r *petSkinImageRepository) GetPetSkinImageByID(ctx context.Context, petSkinImageID uint) (*model.PetSkinImage, error) {
	var petSkinImage model.PetSkinImage

	if err := r.db.WithContext(ctx).
		Model(&model.PetSkinImage{}).
		Where("id = ?", petSkinImageID).
		First(&petSkinImage).Error; err != nil {
		return nil, fmt.Errorf("failed to get pet skin image by id : %w", err)
	}

	return &petSkinImage, nil
}

func (r *petSkinImageRepository) UpdateLabeledPetSkinDiseaseByID(ctx context.Context, petSkinImageID uint, petID uint, userID uint, labeled model.DiseaseType, imageEvidence string) error {
	authorizedPetSkinImageIDQuery := r.db.WithContext(ctx).
		Model(&model.PetSkinImage{}).
		Select("pet_skin_images.id").
		Joins("JOIN pets ON pets.id = pet_skin_images.pet_id AND pets.deleted_at IS NULL").
		Where("pet_skin_images.id = ? AND pet_skin_images.pet_id = ? AND pets.user_id = ?", petSkinImageID, petID, userID).
		Limit(1)

	result := r.db.WithContext(ctx).
		Model(&model.PetSkinImage{}).
		Where("id = (?)", authorizedPetSkinImageIDQuery).
		Updates(map[string]interface{}{
			"labeled":        labeled,
			"image_evidence": imageEvidence,
		})
	if result.Error != nil {
		return fmt.Errorf("failed to update labeled pet skin disease by id : %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return utils.ErrNoRowsUpdated
	}

	return nil
}
