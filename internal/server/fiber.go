package server

import (
	"io"
	"log"
	"os"
	"time"

	// _ "github.com/231031/pet-health/docs"
	"github.com/231031/pethealth-backend/internal/applogger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

func Run() {
	// Load environment variables
	cfg := getAllENV()

	app := fiber.New(fiber.Config{
		AppName:       "PetHealth API",
		CaseSensitive: true,
		ServerHeader:  "Fiber",
		ReadTimeout:   5 * time.Second,
		WriteTimeout:  5 * time.Second,
		IdleTimeout:   5 * time.Second,
	})

	// Middleware
	app.Use(compress.New())
	app.Use(etag.New())
	app.Use(favicon.New())
	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(cors.New(
		cors.Config{
			AllowOrigins: "*",
			AllowMethods: "GET,POST,PUT,PATCH,DELETE",
		},
	))
	app.Use(limiter.New(
		limiter.Config{
			Max: 100,
			LimitReached: func(c *fiber.Ctx) error {
				return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
					"status": "failed",
					"error":  "Too many requests, please try again later.",
				})
			},
		},
	))

	// setup fiber logger
	file := InitLogger(app)
	multiWriter := io.MultiWriter(os.Stdout, file)

	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("error closing log file: %v", err)
		}
	}()

	// setup app logger
	applogger.InitLogtime(multiWriter)

	// setup routes
	Setup(app, cfg)

	err := app.Listen(":" + cfg.BACKEND_PORT)
	if err != nil {
		log.Fatal("failed to start server:", err)
	}
	log.Println("Server is running on port", cfg.BACKEND_PORT)
}
