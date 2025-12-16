package server

import (
	"github.com/231031/pethealth-backend/internal/controller"
	"github.com/231031/pethealth-backend/internal/middleware"
	"github.com/231031/pethealth-backend/internal/repository"
	"github.com/231031/pethealth-backend/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func CreateRoute(router fiber.Router, db *gorm.DB, redisClient *redis.Client, cfg *Cfg) {
	tokenCfg := ConfigGenerateKey(cfg)

	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(redisClient)

	tokenService := service.NewTokenService(tokenRepo, userRepo, tokenCfg)
	authService := service.NewAuthService(userRepo, tokenService)
	authController := controller.NewAuthController(authService)

	userService := service.NewUserService(userRepo)
	userController := controller.NewUserController(userService)

	authMiddlware := middleware.NewAuthMiddleware(tokenService)

	// routing
	authRoute := router.Group("/auth")
	authRoute.Post("/register", authController.CreateUser)
	authRoute.Post("/login", authController.LoginUser)

	userRoute := router.Group("/user", authMiddlware.AuthorizeUser())
	userRoute.Get("/", userController.GetUserAllInfo)
}
