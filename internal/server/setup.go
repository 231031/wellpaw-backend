package server

import (
	"fmt"
	"log"

	"github.com/231031/pethealth-backend/internal/applogger"
	"github.com/231031/pethealth-backend/internal/middleware"
	"github.com/231031/pethealth-backend/internal/migration"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

func Setup(app *fiber.App, cfg *Cfg) {
	// Connect to Postgres
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DB_HOST,
		cfg.DB_PORT,
		cfg.DB_USER,
		cfg.DB_PASSWORD,
		cfg.DB_NAME,
	)
	db, err := ConnectPostgres(dsn)
	if err != nil {
		log.Fatal("failed to connect to Postgres:", err)
	}
	applogger.LogInfo(fmt.Sprintln("connected to Postgres"), serverLog)

	mg := migration.NewMigrationManager(db)
	if err := mg.MigrateToDB(); err != nil {
		log.Fatal("failed to migrate to Postgres:", err)
	}
	applogger.LogInfo("database migration completed", serverLog)

	client, err := ConnectRedis(cfg.REDIS_HOST, cfg.REDIS_PORT, cfg.REDIS_PASSWORD)
	if err != nil {
		applogger.LogError(fmt.Sprintln("failed to connect to Redis:", err), serverLog)
	}
	applogger.LogInfo("connected to Redis", serverLog)

	// helath check
	app.Get("health", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	// Swagger documentation
	app.Get("/swagger/*", swagger.HandlerDefault)

	api := app.Group("/api", middleware.AcceptMiddleware("application/json", "text/plain"))
	CreateRoute(api, db, client, cfg)

}
