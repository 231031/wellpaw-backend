package repository

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/231031/wellpaw-backend/internal/applogger"
	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/231031/wellpaw-backend/internal/utils"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	UpdatePetUpdateNotification(ctx context.Context, id uint, notiUpdatePet bool) error
	GetSubscriptionDetailFromDB(ctx context.Context, userID uint) (*model.User, error)
	GetSubscriptionDetail(ctx context.Context, userID uint) (*model.TierType, *model.SubscriptionStatusType, error)
	SetCurrentSubscriptionDetail(ctx context.Context, userID uint, tier model.TierType, subscriptionStatus model.SubscriptionStatusType) error
	UpdatePaymentMethod(ctx context.Context, id uint, paymentMethodID string) error
	UpdatePasswordByEmail(ctx context.Context, email string, password string) error
	UpdateCustomerID(ctx context.Context, id uint, customerID string) error
	UpdateSubscriptionDetail(ctx context.Context, customerID string, status model.SubscriptionStatusType, tier model.TierType) (*model.User, error)
}

type userRepository struct {
	db          *gorm.DB
	redisClient *redis.Client
	redisTTL    time.Duration
}

func NewUserRepository(db *gorm.DB, redisClient *redis.Client) UserRepository {
	return &userRepository{
		db:          db,
		redisClient: redisClient,
		redisTTL:    15 * 24 * time.Hour,
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
		Select("id", "email", "first_name", "last_name", "customer_id",
			"noti_food", "noti_calendars", "noti_update_pet",
			"profile_free", "food_free", "food_plan_free", "disease_free",
			"tier", "subscription_status").
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
			"noti_food", "noti_calendars", "noti_update_pet",
			"profile_free", "food_free", "food_plan_free", "disease_free",
			"tier", "subscription_status").
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
		Select("id", "email", "first_name", "last_name", "customer_id",
			"noti_food", "noti_calendars", "noti_update_pet",
			"profile_free", "food_free", "food_plan_free", "disease_free",
			"tier", "subscription_status").
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

func (r *userRepository) UpdatePetUpdateNotification(ctx context.Context, id uint, notiUpdatePet bool) error {
	result := r.db.WithContext(ctx).Table("users").Where("id = ?", id).Update("noti_update_pet", notiUpdatePet)
	if result.Error != nil {
		return fmt.Errorf("failed to update notification : %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return utils.ErrNoRowsUpdated
	}

	return nil
}

func (r *userRepository) GetSubscriptionDetailFromDB(ctx context.Context, userID uint) (*model.User, error) {
	var user *model.User
	err := r.db.WithContext(ctx).
		Select("tier", "subscription_status", "profile_free", "food_free", "food_plan_free", "disease_free").Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription detail from db : %w", err)
	}

	return user, nil
}

func (r *userRepository) getRedisKey(userID uint) string {
	return fmt.Sprintf("sub:%d", userID)
}

func (r *userRepository) SetCurrentSubscriptionDetail(ctx context.Context, userID uint, tier model.TierType, status model.SubscriptionStatusType) error {
	key := r.getRedisKey(userID)
	value := fmt.Sprintf("%d:%d", tier, status)
	_, err := r.redisClient.SetArgs(ctx, key, value, redis.SetArgs{
		Get: true,
		TTL: r.redisTTL,
	}).Result()

	if err == redis.Nil {
		return nil
	} else if err != nil {
		applogger.LogError(fmt.Sprintf("redis failed to set subscription detail : %s", err.Error()), repoLog)
		return fmt.Errorf("failed to set subscription detail : %w", err)
	}

	return nil
}

func (r *userRepository) GetSubscriptionDetail(ctx context.Context, userID uint) (*model.TierType, *model.SubscriptionStatusType, error) {
	key := r.getRedisKey(userID)
	value, err := r.redisClient.Get(ctx, key).Result()
	if err != nil {
		applogger.LogError(fmt.Sprintf("redis failed to get subscription detail : %s", err.Error()), repoLog)

		user, err := r.GetSubscriptionDetailFromDB(ctx, userID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get subscription detail : %w", err)
		}

		if err := r.SetCurrentSubscriptionDetail(ctx, userID, user.Tier, user.SubscriptionStatus); err != nil {
			return &user.Tier, &user.SubscriptionStatus, nil
		}

		return &user.Tier, &user.SubscriptionStatus, nil
	}

	result := strings.Split(value, ":")
	tier, err := strconv.Atoi(result[0])
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get subscription detail : %w", err)
	}
	tierType := model.TierType(tier)

	status, err := strconv.Atoi(result[1])
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get subscription detail : %w", err)
	}
	statusType := model.SubscriptionStatusType(status)

	return &tierType, &statusType, nil
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

func (r *userRepository) UpdatePasswordByEmail(ctx context.Context, email string, password string) error {
	result := r.db.WithContext(ctx).Table("users").
		Where("LOWER(email) = LOWER(?)", strings.TrimSpace(email)).
		Update("password", password)
	if result.Error != nil {
		return fmt.Errorf("failed to update password : %w", result.Error)
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

func (r *userRepository) UpdateSubscriptionDetail(ctx context.Context, customerID string, status model.SubscriptionStatusType, tier model.TierType) (*model.User, error) {
	updateData := map[string]interface{}{
		"subscription_status": status,
		"tier":                tier,
	}

	var user *model.User
	result := r.db.WithContext(ctx).
		Model(&model.User{}).
		Clauses(clause.Returning{}).
		Where("customer_id = ?", customerID).
		Updates(updateData).
		Scan(&user)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to update subscription status : %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return user, utils.ErrNoRowsUpdated
	}

	return user, nil
}
