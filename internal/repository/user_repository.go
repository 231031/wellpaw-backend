package repository

import (
	"context"
	"fmt"

	"github.com/231031/pethealth-backend/internal/model"
	"github.com/231031/pethealth-backend/internal/utils"
	"gorm.io/gorm"
)

var (
	repositoryLog = "[REPOSITORY LOGGER]"
)

type UserRepository interface {
	CreateUser(ctx context.Context, u *model.User) error
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	GetUserByID(ctx context.Context, id uint) (*model.User, error)
	GetUserAllInfo(ctx context.Context, id uint) (*model.User, error)
	UpdateUser(ctx context.Context, u *model.User) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) CreateUser(ctx context.Context, u *model.User) error {
	if err := r.db.WithContext(ctx).Create(u).Error; err != nil {
		return fmt.Errorf("failed to create user : %w", err)
	}

	return nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var user *model.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to get user by email : %w", err)
	}

	return user, nil
}

func (r *userRepository) GetUserByID(ctx context.Context, id uint) (*model.User, error) {
	var user *model.User
	err := r.db.WithContext(ctx).
		Select("id", "email", "first_name", "last_name",
			"noti_food", "noti_calendars",
			"profile_free", "food_free", "food_plan_free", "bcs_free", "disease_free",
			"payment_plan").
		First(&user, id).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id : %w", err)
	}

	return user, nil
}

func (r *userRepository) GetUserAllInfo(ctx context.Context, id uint) (*model.User, error) {
	var user *model.User

	// all info in dashboard page
	err := r.db.WithContext(ctx).
		Preload("Pets").
		Select("id", "email", "first_name", "last_name",
			"noti_food", "noti_calendars",
			"profile_free", "food_free", "food_plan_free", "bcs_free", "disease_free",
			"payment_plan").
		First(&user, id).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id : %w", err)
	}

	return user, nil
}

func (r *userRepository) UpdateUser(ctx context.Context, u *model.User) error {
	result := r.db.WithContext(ctx).Updates(u)
	if result.Error != nil {
		return fmt.Errorf("failed to update user : %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return utils.ErrNoRowsUpdated
	}

	return nil
}
