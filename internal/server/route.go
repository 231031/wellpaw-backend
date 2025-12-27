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
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(redisClient)

	tokenCfg := ConfigGenerateKey(cfg)
	tokenService := service.NewTokenService(tokenRepo, userRepo, tokenCfg)

	googleOauthConfig := ConfigGoogleOauthConfig(cfg)
	authService := service.NewAuthService(userRepo, tokenService, googleOauthConfig)
	authController := controller.NewAuthController(authService)

	userService := service.NewUserService(userRepo)
	userController := controller.NewUserController(userService)

	authMiddlware := middleware.NewAuthMiddleware(tokenService)

	// routing
	authRoute := router.Group("/auth")
	authRoute.Post("/register", authController.CreateUser)
	authRoute.Post("/login", authController.LoginUser)
	authRoute.Post("/login/google", authController.LoginUserWithGoogle)
	authRoute.Post("/refreshtoken", authController.RefreshToken)

	userRoute := router.Group("/user", authMiddlware.AuthorizeUser())
	userRoute.Get("/", userController.GetUserAllInfo)
}
