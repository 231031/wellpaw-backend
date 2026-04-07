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

	fcmClient, err := connectFirebaseMessaging()
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
	app.Get("/privacy", privacyHandler)
	app.Get("/deletepolicy", deletePolicyHandler)

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
			<h2>Business Information</h2>
			<p><strong>Business Name:</strong> WellPaw (Individual Developer)</p>
			<p><strong>Location:</strong> Thailand</p>
			<p><strong>Description:</strong> WellPaw is a digital pet health care service application that helps pet owners track nutrition, manage pet weight, and monitor potential skin health issues.</p>
		</div>

		<div class="section">
			<h2>Pricing & Products</h2>
			<p>WellPaw provides digital services through the following plans. Payments are processed securely via Stripe and major credit cards are accepted.</p>

			<ul>
				<li><strong>Free Trial:</strong> Provides temporary access to limited features for evaluation.</li>
				<li><strong>Monthly Subscription:</strong> ฿69.00 THB per month.</li>
				<li><strong>Yearly Subscription:</strong> ฿690.00 THB per year.</li>
			</ul>

			<p>Subscriptions automatically renew at the end of each billing period unless cancelled before the renewal date.</p>
		</div>

		<div class="section">
			<h2>Refund & Cancellation Policy</h2>

			<p>Subscriptions can be cancelled at any time through the application settings.</p>

			<p>When a subscription is cancelled, access to premium features will end immediately.</p>

			<p>Because our services are digital products delivered instantly, 
			<strong>we do not offer refunds for completed payments or partially used subscription periods.</strong></p>

			<p>Users may continue using any remaining active subscription time until cancellation takes effect if applicable.</p>
		</div>

		<div class="section">
			<h2>Terms of Service</h2>

			<p>By using WellPaw, you agree to use the application only for personal pet care tracking and management purposes.</p>

			<p>WellPaw provides informational tools and does not replace professional veterinary advice or diagnosis.</p>

			<p>Subscriptions provide access to premium features during the active billing period and renew automatically unless cancelled.</p>
		</div>

		<div class="section">
			<h2>Privacy Policy</h2>

			<p>WellPaw collects limited information necessary to operate the service, such as email address, account data, and pet-related information entered by users.</p>

			<p>Payment information is securely processed by Stripe. WellPaw does not store credit card numbers on its servers.</p>

			<p>User information is used only to provide and improve the service and is not sold to third parties.</p>
		</div>

		<div class="section">
			<h2>Customer Support</h2>

			<p>If you have questions about billing, subscriptions, or your account, please contact:</p>

			<p><strong>Email:</strong> mybile.e70e@gmail.com</p>
		</div>

	</body>
	</html>
	`

	c.Type("html")
	return c.SendString(htmlContent)
}

func privacyHandler(c *fiber.Ctx) error {
	html := `
		<!DOCTYPE html>
		<html lang="en">
		<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>WellPaw - Privacy Policy</title>
		<style>
		body { font-family: Arial; max-width: 800px; margin: 40px auto; line-height: 1.6; color: #333; }
		h1, h2 { color: #444; }
		</style>
		</head>
		<body>

		<h1>Privacy Policy</h1>

		<p><strong>Effective Date:</strong> April 7, 2026</p>

		<p>
		WellPaw we operates the WellPaw mobile application.
		This Privacy Policy explains how we collect, use, and protect your information.
		</p>

		<h2>1. Information We Collect</h2>
		<ul>
		<li>Email address</li>
		<li>Full name</li>
		<li>Pet-related data (e.g., weight, nutrition, health tracking)</li>
		<li>Authentication data (Email/Password or Google OAuth)</li>
		<li>Basic device/log information (for app performance and security)</li>
		</ul>

		<h2>2. How We Use Your Information</h2>
		<ul>
		<li>Provide and maintain app functionality</li>
		<li>Authenticate users securely</li>
		<li>Improve features and user experience</li>
		<li>Ensure security and prevent misuse</li>
		</ul>

		<h2>3. Legal Basis</h2>
		<p>
		We process your data based on user consent and for providing the requested service.
		</p>

		<h2>4. Data Retention</h2>
		<p>
		We retain user data only as long as necessary to provide services.
		Deleted accounts are permanently removed within 30 days unless required by law.
		</p>

		<h2>5. Data Sharing</h2>
		<p>
		We do not sell your personal data. We may share data with trusted services such as:
		</p>
		<ul>
		<li>Google (for authentication)</li>
		</ul>

		<h2>6. Data Security</h2>
		<p>
		We implement reasonable technical and organizational measures to protect your data.
		Passwords are encrypted and access is restricted.
		</p>

		<h2>7. Your Rights</h2>
		<ul>
		<li>Request access to your data</li>
		<li>Request correction or deletion</li>
		<li>Withdraw consent at any time</li>
		</ul>

		<h2>8. Children's Privacy</h2>
		<p>
		This app is not intended for children under 13 years of age.
		</p>

		<h2>9. Changes to This Policy</h2>
		<p>
		We may update this Privacy Policy. Changes will be posted on this page.
		</p>

		<h2>10. Contact</h2>
		<p>
		Developer: WellPaw Team<br>
		Country: Thailand<br>
		Email: mybile.e70e@gmail.com
		</p>

		</body>
		</html>
	`
	c.Type("html")
	return c.SendString(html)
}

func deletePolicyHandler(c *fiber.Ctx) error {
	html := `
		<!DOCTYPE html>
		<html lang="en">
		<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>WellPaw - Account Deletion Policy</title>
		<style>
		body { font-family: Arial; max-width: 800px; margin: 40px auto; line-height: 1.6; color: #333; }
		h1, h2 { color: #444; }
		.highlight { background: #f8f8f8; padding: 10px; border-left: 4px solid #d33; }
		</style>
		</head>
		<body>

		<h1>Account Deletion Policy</h1>

		<p><strong>Effective Date:</strong> April 7, 2026</p>

		<h2>How to Request Deletion</h2>
		<p>
		Users can request account deletion directly within the app (if available) or by contacting us via email.
		</p>

		<p><strong>Email:</strong> mybile.e70e@gmail.com</p>

		<h2>Required Information</h2>
		<ul>
		<li>Registered email address</li>
		</ul>

		<h2>What Will Be Deleted</h2>
		<ul>
		<li>User account information</li>
		<li>Authentication data</li>
		<li>All pet-related records</li>
		</ul>

		<h2>Data Retention</h2>
		<p>
		Some minimal data may be retained for up to 30 days for legal and security purposes.
		</p>

		<h2>Processing Time</h2>
		<p>
		Requests are processed within 3–7 business days.
		</p>

		<h2>Important</h2>
		<ul>
		<li>Deletion is permanent</li>
		<li>Data cannot be recovered</li>
		</ul>

		<h2>Contact</h2>
		<p>Email: mybile.e70e@gmail.com</p>

		</body>
		</html>
	`
	c.Type("html")
	return c.SendString(html)
}
