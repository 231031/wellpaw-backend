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
	firebaseStorage, err = connectFirebaseStorage(cfg)
	if err != nil {
		applogger.LogError(fmt.Sprintln("failed to connect to Firebase Storage:", err), serverLog)
	} else {
		applogger.LogInfo("connected to Firebase Storage", serverLog)
	}

	fcmClient, err := connectFirebaseMessaging(cfg)
	if err != nil {
		applogger.LogError(fmt.Sprintln("failed to connect to Firebase Messaging Cloud:", err), serverLog)
	} else {
		applogger.LogInfo("connected to Firebase Messaging Cloud", serverLog)
	}

	// helath check
	app.Get("health", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	// Swagger documentation
	app.Get("/swagger/*", swagger.HandlerDefault)

	// policy page
	app.Get("/policy", policyHandler)

	router := app.Group("/api", middleware.AcceptMiddleware("application/json", "text/plain", "image/*"))
	CreateRoute(router, db, redisClient, geminiClient, stripeClient, firebaseStorage, fcmClient, cfg)

}

func policyHandler(c *fiber.Ctx) error {
	htmlContent := `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>WellPaw - Business Information & Policies</title>
		<style>
			body { font-family: Arial, sans-serif; line-height: 1.6; max-width: 800px; margin: 40px auto; padding: 20px; color: #333; }
			h1, h2 { color: #4A4A4A; }
			.section { margin-bottom: 30px; }
		</style>
	</head>
	<body>
		<h1>WellPaw - Business Information & Policies</h1>
		
		<div class="section">
			<h2>Business Name & Description</h2>
			<p><strong>Business Name:</strong> Kanlayaphat Prakobwaitayakij (Individual)</p>
			<p><strong>Description:</strong> WellPaw is a health care service application for pets. It helps owners track pet nutrition, manage weight, and identify potential skin diseases.</p>
		</div>

		<div class="section">
			<h2>Pricing & Products</h2>
			<p>We offer digital access to our platform via the following plans. We accept major credit cards for all transactions.</p>
			<ul>
				<li><strong>Free Trial:</strong> Access to basic features for a limited time.</li>
				<li><strong>Monthly Plan:</strong> ฿69.00 THB per month.</li>
				<li><strong>Yearly Plan:</strong> ฿690.00 THB per year.</li>
			</ul>
		</div>

		<div class="section">
			<h2>Refund & Cancellation Policy</h2>
			<p>Subscriptions can be canceled at any time directly through the application settings.</p>
			<p>Because our products are digital services, <strong>we do not offer cash refunds or returns</strong> to your original payment method for completed transactions, one-time purchases, or canceled subscription periods.</p>
			<p>If you choose to cancel your subscription before the end of your billing cycle, your premium access will end immediately. However, the remaining unused value of your current billing cycle will be automatically converted into account credit. This credit is securely managed by our payment processor and will be automatically applied to any of your future purchases within the application.</p>
		</div>

		<div class="section">
			<h2>Customer Support</h2>
			<p>If you have any questions or need assistance with your account or billing, please contact us at:</p>
			<p><strong>Email:</strong> mybile.e70e@gmail.com</p> 
		</div>
	</body>
	</html>
	`

	// c.Type("html") automatically sets the "Content-Type: text/html" header
	c.Type("html")

	// c.SendString returns the HTML string to the browser
	return c.SendString(htmlContent)
}
