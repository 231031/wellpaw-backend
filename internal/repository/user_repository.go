package repository

import (
	"context"
	"fmt"

	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/231031/wellpaw-backend/internal/utils"
	"gorm.io/gorm"
)

// var (
// 	repositoryLog = "[REPOSITORY LOGGER]"
// )

type UserRepository interface {
	CreateUser(ctx context.Context, u *model.User) error
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	GetUserByID(ctx context.Context, id uint) (*model.User, error)
	GetUserIdDetailByID(ctx context.Context, id uint) (*model.User, error)
	GetUserAllInfo(ctx context.Context, id uint) (*model.User, error)
	UpdateUser(ctx context.Context, u *model.User) error
	UpdateFoodNotification(ctx context.Context, id uint, notiFood bool) error
	UpdateCalendarNotification(ctx context.Context, id uint, notiCalendar bool) error
	UpdatePaymentMethod(ctx context.Context, id uint, paymentMethodID string) error
	UpdateCustomerID(ctx context.Context, id uint, customerID string) error
	UpdateSubscriptionDetail(ctx context.Context, email string, status model.SubscriptionStatusType, tier model.TierType) error
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

func (r *userRepository) GetUserIdDetailByID(ctx context.Context, id uint) (*model.User, error) {
	var user *model.User
	err := r.db.WithContext(ctx).
		Select("id", "email", "first_name", "last_name", "device_token",
			"payment_method_id", "customer_id",
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

func (r *userRepository) UpdateFoodNotification(ctx context.Context, id uint, notiFood bool) error {
	result := r.db.WithContext(ctx).Table("users").Where("id = ?", id).Update("noti_food", notiFood)
	if result.Error != nil {
		return fmt.Errorf("failed to update notification : %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return utils.ErrNoRowsUpdated
	}

	return nil
}

func (r *userRepository) UpdateCalendarNotification(ctx context.Context, id uint, notiCalendar bool) error {
	result := r.db.WithContext(ctx).Table("users").Where("id = ?", id).Update("noti_calendars", notiCalendar)
	if result.Error != nil {
		return fmt.Errorf("failed to update notification : %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return utils.ErrNoRowsUpdated
	}

	return nil
}

func (r *userRepository) UpdatePaymentMethod(ctx context.Context, id uint, paymentMethodID string) error {
	result := r.db.WithContext(ctx).Table("users").Where("id = ?", id).Update("payment_method_id", paymentMethodID)
	if result.Error != nil {
		return fmt.Errorf("failed to update payment method : %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return utils.ErrNoRowsUpdated
	}

	return nil
}

func (r *userRepository) UpdateCustomerID(ctx context.Context, id uint, customerID string) error {
	result := r.db.WithContext(ctx).Table("users").Where("id = ?", id).Update("customer_id", customerID)
	if result.Error != nil {
		return fmt.Errorf("failed to update customer id : %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return utils.ErrNoRowsUpdated
	}

	return nil
}

func (r *userRepository) UpdateSubscriptionDetail(ctx context.Context, customerID string, status model.SubscriptionStatusType, tier model.TierType) error {
	result := r.db.WithContext(ctx).Table("users").Where("customer_id = ?", customerID).Updates(map[string]interface{}{
		"subscription_status": status,
		"tier":                tier,
	})
	if result.Error != nil {
		return fmt.Errorf("failed to update subscription status : %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return utils.ErrNoRowsUpdated
	}

	return nil
}
