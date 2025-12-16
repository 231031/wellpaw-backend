package server

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectPostgres(dsn string) (*gorm.DB, error) {
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
