package server

import (
	"github.com/231031/wellpaw-backend/internal/controller"
	"github.com/231031/wellpaw-backend/internal/middleware"
	"github.com/231031/wellpaw-backend/internal/repository"
	"github.com/231031/wellpaw-backend/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/google/generative-ai-go/genai"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func RouteAuth(router fiber.Router, authController controller.AuthController) {
	authRoute := router.Group("/auth")
	authRoute.Post("/register", authController.CreateUser)
	authRoute.Post("/login", authController.LoginUser)
	authRoute.Post("/login/google", authController.LoginUserWithGoogle)
	authRoute.Post("/refreshtoken", authController.RefreshToken)
}

func RouteUser(router fiber.Router, userController controller.UserController, authMiddleware middleware.AuthMiddleware) {
	userRoute := router.Group("/user", authMiddleware.AuthorizeUser())
	userRoute.Get("/", userController.GetUserAllInfo)
	userRoute.Get("/notification/food", userController.ManageFoodNotification)
	userRoute.Get("/notification/calendar", userController.ManageCalendarNotification)
}

func RouteOcr(router fiber.Router, ocrController controller.OcrController, authMiddleware middleware.AuthMiddleware) {
	ocrRoute := router.Group("/ocr", authMiddleware.AuthorizeUser())
	ocrRoute.Post("/request", ocrController.ProcessOcrRequest)
}

func CreateRoute(router fiber.Router, db *gorm.DB, redisClient *redis.Client, geminiClient *genai.Client, cfg *Cfg) {
	tokenCfg := ConfigGenerateKey(cfg)
	googleOauthConfig := ConfigGoogleOauthConfig(cfg)

	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(redisClient)

	tokenService := service.NewTokenService(tokenRepo, userRepo, tokenCfg)
	authMiddlware := middleware.NewAuthMiddleware(tokenService)

	// routing
	authService := service.NewAuthService(userRepo, tokenService, googleOauthConfig)
	authController := controller.NewAuthController(authService)
	RouteAuth(router, authController)

	userService := service.NewUserService(userRepo)
	userController := controller.NewUserController(userService)
	RouteUser(router, userController, authMiddlware)

	ocrService := service.NewOcrService(geminiClient)
	ocrController := controller.NewOcrController(ocrService)
	RouteOcr(router, ocrController, authMiddlware)
}
