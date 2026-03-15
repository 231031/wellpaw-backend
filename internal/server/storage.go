package server

import (
	"context"
	"fmt"
	"strings"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/redis/go-redis/v9"
	"google.golang.org/api/option"
	gcpstorage "google.golang.org/api/storage/v1"
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

func connectFirebaseStorage(cfg *Cfg) (*model.FirebaseStorage, error) {
	if cfg == nil {
		return nil, fmt.Errorf("firebase storage config is nil")
	}

	bucketName := strings.TrimSpace(cfg.FIREBASE_STORAGE_BUCKET)
	if bucketName == "" {
		return nil, fmt.Errorf("firebase storage bucket is not configured")
	}

	ctx := context.Background()
	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	service, err := gcpstorage.NewService(ctxTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize firebase storage client: %w", err)
	}

	if _, err := service.Buckets.Get(bucketName).Context(ctxTimeout).Do(); err != nil {
		return nil, fmt.Errorf("failed to access firebase storage bucket %s: %w", bucketName, err)
	}

	return &model.FirebaseStorage{
		Objects:    service.Objects,
		BucketName: bucketName,
	}, nil
}

func connectFirebaseMessaging(cfg *Cfg) (*messaging.Client, error) {
	ctx := context.Background()
	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	opt := option.WithCredentialsFile(cfg.GOOGLE_APPLICATION_CREDENTIALS)
	app, err := firebase.NewApp(ctxTimeout, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize create new firebase app: %w", err)
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize firebase messaging cloud: %w", err)
	}

	return client, nil
}
