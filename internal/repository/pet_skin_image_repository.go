package repository

import (
	"context"
	"fmt"

	"github.com/231031/wellpaw-backend/internal/model"
	"gorm.io/gorm"
)

type PetSkinImageRepository interface {
	CreatePetSkinImage(ctx context.Context, petSkinImage *model.PetSkinImage) error
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
