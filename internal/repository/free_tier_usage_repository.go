package repository

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/231031/wellpaw-backend/internal/applogger"
	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/redis/go-redis/v9"
)

type FreeTierUsageRepository interface {
	// GetFreeTierUsage(ctx context.Context, userID uint) (*model.FreeTierUsage, error)
	SetFreeTierUsage(ctx context.Context, userID uint, usage *model.FreeTierUsage) error
	GetDeleteFreeTierUsage(ctx context.Context, userID uint) (*model.FreeTierUsage, error)
}

type freeTierUsageRepository struct {
	redisClient *redis.Client
	redisTTL    time.Duration
}

func NewFreeTierUsageRepository(redisClient *redis.Client) FreeTierUsageRepository {
	return &freeTierUsageRepository{
		redisClient: redisClient,
		redisTTL:    15 * 24 * time.Hour,
	}
}

func (r *freeTierUsageRepository) GetFreeTierUsage(ctx context.Context, userID uint) (*model.FreeTierUsage, error) {
	value, err := r.redisClient.Get(ctx, r.getRedisKey(userID)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, redis.Nil
		}

		applogger.LogError(fmt.Sprintf("failed to get free tier usage from redis: %s", err.Error()), repoLog)
		return nil, fmt.Errorf("failed to get free tier usage from redis: %w", err)
	}

	usage, err := parseFreeTierUsage(value)
	if err != nil {
		return nil, err
	}

	return usage, nil
}

func (r *freeTierUsageRepository) SetFreeTierUsage(ctx context.Context, userID uint, usage *model.FreeTierUsage) error {
	key := r.getRedisKey(userID)
	value := r.formatValue(usage)
	_, err := r.redisClient.SetArgs(ctx, key, value, redis.SetArgs{
		Get: true,
		TTL: r.redisTTL,
	}).Result()
	if err != nil {
		applogger.LogError(fmt.Sprintf("failed to set free tier usage to redis: %s", err.Error()), repoLog)
		return fmt.Errorf("failed to set free tier usage to redis: %w", err)
	}

	return nil
}

func (r *freeTierUsageRepository) GetDeleteFreeTierUsage(ctx context.Context, userID uint) (*model.FreeTierUsage, error) {
	value, err := r.redisClient.GetDel(ctx, r.getRedisKey(userID)).Result()
	if err != nil {
		applogger.LogError(fmt.Sprintf("failed to delete free tier usage from redis: %s", err.Error()), repoLog)
		return nil, fmt.Errorf("failed to delete free tier usage from redis: %w", err)
	}

	usage, err := parseFreeTierUsage(value)
	if err != nil {
		return nil, err
	}

	return usage, nil
}

func (r *freeTierUsageRepository) getRedisKey(userID uint) string {
	return fmt.Sprintf("free:%d", userID)
}

func (r *freeTierUsageRepository) formatValue(usage *model.FreeTierUsage) string {
	return fmt.Sprintf("profile:%d:food:%d:foodplan:%d:disease:%d", usage.ProfileFree, usage.FoodFree, usage.FoodPlanFree, usage.DiseaseFree)
}

func parseFreeTierUsage(value string) (*model.FreeTierUsage, error) {
	splitValue := strings.Split(value, ":")
	if len(splitValue) != 8 || splitValue[0] != "profile" || splitValue[2] != "food" || splitValue[4] != "foodplan" || splitValue[6] != "disease" {
		return nil, fmt.Errorf("invalid free tier usage format")
	}

	profileFree, err := strconv.Atoi(splitValue[1])
	if err != nil {
		return nil, fmt.Errorf("invalid profile free format: %w", err)
	}

	foodFree, err := strconv.Atoi(splitValue[3])
	if err != nil {
		return nil, fmt.Errorf("invalid food free format: %w", err)
	}

	foodPlanFree, err := strconv.Atoi(splitValue[5])
	if err != nil {
		return nil, fmt.Errorf("invalid food plan free format: %w", err)
	}

	diseaseFree, err := strconv.Atoi(splitValue[7])
	if err != nil {
		return nil, fmt.Errorf("invalid disease free format: %w", err)
	}

	return &model.FreeTierUsage{
		ProfileFree:  profileFree,
		FoodFree:     foodFree,
		FoodPlanFree: foodPlanFree,
		DiseaseFree:  diseaseFree,
	}, nil
}
