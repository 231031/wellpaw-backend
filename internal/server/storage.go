package server

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectPostgres(cfg *Cfg) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DB_HOST,
		cfg.DB_PORT,
		cfg.DB_USER,
		cfg.DB_PASSWORD,
		cfg.DB_NAME,
	)
	db, err := gorm.Open(
		postgres.Open(dsn), &gorm.Config{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func connectRedis(host, port, password string) (*redis.Client, error) {
	addr := fmt.Sprintf("%s:%s", host, port)
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	ctx := context.Background()
	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return client, client.Ping(ctxTimeout).Err()
}
