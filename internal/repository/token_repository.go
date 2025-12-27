package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/231031/wellpaw-backend/internal/applogger"
	"github.com/redis/go-redis/v9"
)

var (
	repoLog = "[REPOSITORY LOGGER]"
)

type TokenRepository interface {
	SetRefreshToken(ctx context.Context, key string, value string, expIn time.Duration) error
	GetandDelRefreshToken(ctx context.Context, key string) (string, error)
	DeleteRefreshToken(ctx context.Context, key string) error
}

type tokenRepository struct {
	client *redis.Client
}

func NewTokenRepository(c *redis.Client) TokenRepository {
	return &tokenRepository{
		client: c,
	}
}

func (r *tokenRepository) SetRefreshToken(ctx context.Context, key string, value string, expIn time.Duration) error {
	err := r.client.Set(ctx, key, value, expIn).Err()
	if err != nil {
		applogger.LogError(fmt.Sprintln("failed to cache refresh token on redis", err), repoLog)
		return nil
	}

	return nil
}

func (r *tokenRepository) GetandDelRefreshToken(ctx context.Context, key string) (string, error) {
	val, err := r.client.GetDel(ctx, key).Result()
	if err != nil {
		applogger.LogError(fmt.Sprintln("failed to get and del refresh token on redis", err), repoLog)
		return "", err
	}

	return val, nil
}

func (r *tokenRepository) DeleteRefreshToken(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		applogger.LogError(fmt.Sprintln("failed to del refresh token on redis", err), repoLog)
		return err
	}

	return nil
}
