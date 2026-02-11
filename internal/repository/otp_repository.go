package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/231031/wellpaw-backend/internal/applogger"
	"github.com/redis/go-redis/v9"
)

var (
	ErrOTPNotFound = errors.New("otp not found or expired")
)

type OTPRepository interface {
	SetOTP(ctx context.Context, email string, otp string, expIn time.Duration) error
	GetOTP(ctx context.Context, email string) (string, error)
	DeleteOTP(ctx context.Context, email string) error
}

type otpRepository struct {
	client *redis.Client
}

func NewOTPRepository(client *redis.Client) OTPRepository {
	return &otpRepository{
		client: client,
	}
}

func (r *otpRepository) getOTPKey(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func (r *otpRepository) SetOTP(ctx context.Context, email string, otp string, expIn time.Duration) error {
	key := r.getOTPKey(email)
	if err := r.client.Set(ctx, key, otp, expIn).Err(); err != nil {
		applogger.LogError(fmt.Sprintf("failed to set otp to redis: %s", err.Error()), repoLog)
		return fmt.Errorf("failed to set otp to redis: %w", err)
	}

	return nil
}

func (r *otpRepository) GetOTP(ctx context.Context, email string) (string, error) {
	key := r.getOTPKey(email)
	otp, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", ErrOTPNotFound
	}
	if err != nil {
		applogger.LogError(fmt.Sprintf("failed to get otp from redis: %s", err.Error()), repoLog)
		return "", fmt.Errorf("failed to get otp from redis: %w", err)
	}

	return otp, nil
}

func (r *otpRepository) DeleteOTP(ctx context.Context, email string) error {
	key := r.getOTPKey(email)
	result, err := r.client.Del(ctx, key).Result()
	if err != nil {
		applogger.LogError(fmt.Sprintf("failed to delete otp from redis: %s", err.Error()), repoLog)
		return fmt.Errorf("failed to delete otp from redis: %w", err)
	}

	if result == 0 {
		return ErrOTPNotFound
	}

	return nil
}
