package server

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/231031/wellpaw-backend/internal/applogger"
	"github.com/231031/wellpaw-backend/internal/middleware"
	"github.com/231031/wellpaw-backend/internal/migration"
	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	"github.com/google/generative-ai-go/genai"
	"github.com/stripe/stripe-go/v84"
	"google.golang.org/api/option"
)

func setupGeminiClient(apiKey string) (*genai.Client, error) {
	ctx := context.Background()
	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	geminiClient, err := genai.NewClient(ctxTimeout, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Google AI: %v", err)
	}

	return geminiClient, nil
}

func Setup(app *fiber.App, cfg *Cfg) {
	// Connect to Postgres
	db, err := ConnectPostgres(cfg)
	if err != nil {
		log.Fatal("failed to connect to Postgres:", err)
	}
	applogger.LogInfo(fmt.Sprintln("connected to Postgres"), serverLog)

	// Migrate to Postgres
	mg := migration.NewMigrationManager(db)
	if err := mg.MigrateToDB(); err != nil {
		applogger.LogError(fmt.Sprintln("failed to migrate to Postgres:", err), serverLog)
	}
	applogger.LogInfo("database migration completed", serverLog)

	// connect Redis
	redisClient, err := connectRedis(cfg.REDIS_HOST, cfg.REDIS_PORT, cfg.REDIS_PASSWORD)
	if err != nil {
		applogger.LogError(fmt.Sprintln("failed to connect to Redis:", err), serverLog)
	}
	applogger.LogInfo("connected to Redis", serverLog)

	// connect Google AI
	geminiClient, err := setupGeminiClient(cfg.GEMINI_API_KEY)
	if err != nil {
		applogger.LogError(fmt.Sprintln("failed to connect to Google AI:", err), serverLog)
	}
	applogger.LogInfo("connected to Google AI", serverLog)

	// connect Stripe
	stripeClient := stripe.NewClient(cfg.STRIPE_API_KEY)
	if err != nil {
		applogger.LogError(fmt.Sprintln("failed to connect to Stripe:", err), serverLog)
	}
	applogger.LogInfo("connected to Stripe", serverLog)

	var firebaseStorage *model.FirebaseStorage

	// connect Firebase Storage (only if configured)
	if cfg.FIREBASE_STORAGE_BUCKET != "" {
		firebaseStorage, err = connectFirebaseStorage(cfg)
		if err != nil {
			applogger.LogError(fmt.Sprintln("failed to connect to Firebase Storage:", err), serverLog)
		} else {
			applogger.LogInfo("connected to Firebase Storage", serverLog)
		}
	}

	// helath check
	app.Get("health", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	// Swagger documentation
	app.Get("/swagger/*", swagger.HandlerDefault)

	router := app.Group("/api", middleware.AcceptMiddleware("application/json", "text/plain", "image/*"))
	CreateRoute(router, db, redisClient, geminiClient, stripeClient, firebaseStorage, cfg)

}
