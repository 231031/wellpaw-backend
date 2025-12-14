package server

import (
	"github.com/231031/pethealth-backend/internal/controller"
	"github.com/231031/pethealth-backend/internal/repository"
	"github.com/231031/pethealth-backend/internal/service"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func CreateRoute(router fiber.Router, db *gorm.DB, cfg *Cfg) {
	userRepo := repository.NewUserRepository(db)

	authService := service.NewAuthService(userRepo)
	authController := controller.NewAuthController(authService)

	authRoute := router.Group("/auth")
	authRoute.Post("/register", authController.CreateUser)
	authRoute.Post("/login", authController.LoginUser)
}
