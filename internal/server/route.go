package server

import (
	"github.com/231031/wellpaw-backend/internal/controller"
	"github.com/231031/wellpaw-backend/internal/middleware"
	"github.com/231031/wellpaw-backend/internal/repository"
	"github.com/231031/wellpaw-backend/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/google/generative-ai-go/genai"
	"github.com/redis/go-redis/v9"
	"github.com/stripe/stripe-go/v84"
	"gorm.io/gorm"
)

func RouteAuth(router fiber.Router, authController controller.AuthController, userController controller.UserController) {
	authRoute := router.Group("/auth")
	authRoute.Post("/register", authController.CreateUser)
	authRoute.Post("/login", authController.LoginUser)
	authRoute.Post("/login/google", authController.LoginUserWithGoogle)
	authRoute.Post("/refreshtoken", authController.RefreshToken)
	authRoute.Post("/otp", authController.RequestOTP)
	authRoute.Post("/resetpassword", authController.ResetPassword)
}

func RouteUser(router fiber.Router, userController controller.UserController, authMiddleware middleware.AuthMiddleware) {
	userRoute := router.Group("/user", authMiddleware.AuthorizeUser())
	userRoute.Get("/", userController.GetUserAllInfo)

	userRoute.Get("/notification/food", userController.ManageFoodNotification)
	userRoute.Get("/notification/calendar", userController.ManageCalendarNotification)

	paymentUserRoute := userRoute.Group("/payment")
	paymentUserRoute.Patch("/paymentmethod", userController.UpdatePaymentMethod)

	subUserRoute := userRoute.Group("/subscription")
	subUserRoute.Get("/", userController.GetAllSubscriptionsPlan)
	subUserRoute.Get("/history", userController.GetAllSubscriptionsByCustomerID)
	subUserRoute.Get("/paymentintent/:payment_intent_id", userController.GetPaymentIntentByID)

	subUserRoute.Post("/start", userController.StartSubscription)

	subUserRoute.Patch("/update", userController.UpdateSubscription)
	subUserRoute.Get("/cancel/:subscription_id", userController.CancelSubscription)
}

func RoutePet(router fiber.Router, petController controller.PetController, authMiddleware middleware.AuthMiddleware) {
	router.Get("/pets", authMiddleware.AuthorizeUser(), petController.GetPetsByUserID)

	petRoute := router.Group("/pet", authMiddleware.AuthorizeUser())
	petRoute.Post("/", petController.CreateNewPet)
	petRoute.Get("/analysis/:pet_id", petController.GetPetAnalysisByPetID)
	petRoute.Put("/info", petController.UpdatePetInfo)
	petRoute.Post("/detail", petController.UpdatePetDetail)
	petRoute.Delete("/:pet_id", petController.SoftDeletePet)
}

func RouteFood(router fiber.Router, foodController controller.FoodController, authMiddleware middleware.AuthMiddleware) {
	router.Get("/foods", authMiddleware.AuthorizeUser(), foodController.GetFoodsByUserID)
	router.Get("/foods/:food_type", authMiddleware.AuthorizeUser(), foodController.GetFoodsByFoodType)

	foodRoute := router.Group("/food", authMiddleware.AuthorizeUser())
	foodRoute.Post("/", foodController.CreateFood)
	foodRoute.Post("/quantity", foodController.CreateNewFoodQuantity)
	foodRoute.Patch("/", foodController.UpdateFoodDetail)
	foodRoute.Delete("/:food_id", foodController.SoftDeleteFood)
}

func RoutePetFoodPlan(router fiber.Router, petFoodPlanController controller.PetFoodPlanController, authMiddleware middleware.AuthMiddleware) {
	petFoodPlanRoute := router.Group("/foodplan", authMiddleware.AuthorizeUser())
	petFoodPlanRoute.Post("/calculate", petFoodPlanController.CalculatePetFoodPlan)
	petFoodPlanRoute.Post("/", petFoodPlanController.CreatePetFoodPlan)
	petFoodPlanRoute.Put("/amount", petFoodPlanController.UpdateFeedingAmountFromUser)
	petFoodPlanRoute.Get("/:pet_id", petFoodPlanController.GetLastestActivePlanDetailByPet)
}

func RouteWebhook(router fiber.Router, webhookController controller.WebhookController) {
	webhookRoute := router.Group("/webhook")
	webhookRoute.Post("/subscription", webhookController.HandleSubscriptionUpdated)
}

func RouteOcr(router fiber.Router, ocrController controller.OcrController, authMiddleware middleware.AuthMiddleware) {
	ocrRoute := router.Group("/ocr", authMiddleware.AuthorizeUser())
	ocrRoute.Post("/request", ocrController.ProcessOcrRequest)
}

func CreateRoute(router fiber.Router, db *gorm.DB, redisClient *redis.Client, geminiClient *genai.Client, stripeClient *stripe.Client, cfg *Cfg) {
	tokenCfg := ConfigGenerateKey(cfg)
	googleOauthConfig := ConfigGoogleOauthConfig(cfg)

	userRepo := repository.NewUserRepository(db, redisClient)
	tokenRepo := repository.NewTokenRepository(redisClient)
	otpRepo := repository.NewOTPRepository(redisClient)

	tokenService := service.NewTokenService(tokenRepo, userRepo, tokenCfg)
	emailService := service.NewEmailService(cfg.MAILJET_API_KEY, cfg.MAILJET_API_SECRET, cfg.MAILJET_SENDER_EMAIL, cfg.MAILJET_SENDER_NAME)
	otpService := service.NewOTPService(otpRepo, emailService)
	authMiddlware := middleware.NewAuthMiddleware(tokenService)

	paymentService := service.NewPaymentService(stripeClient)
	webhookService := service.NewWebhookService(userRepo)

	userService := service.NewUserService(userRepo, paymentService)
	userController := controller.NewUserController(userService)

	energyReqService := service.NewEnergyRequirementService()
	nutritientReqService := service.NewNutritientRequirementService()
	calculationService := service.NewCalculationService(energyReqService, nutritientReqService)

	foodRepo := repository.NewFoodRepository(db)
	foodService := service.NewFoodService(calculationService, foodRepo)
	foodController := controller.NewFoodController(foodService)

	petRepo := repository.NewPetRepository(db)
	petFoodPlanRepo := repository.NewPetFoodPlanRepository(db)

	petFoodPlanService := service.NewPetFoodPlanService(calculationService, petFoodPlanRepo, petRepo, foodRepo)
	petFoodPlanController := controller.NewPetFoodPlanController(petFoodPlanService)

	petService := service.NewPetService(calculationService, petRepo, petFoodPlanRepo)
	petController := controller.NewPetController(petService)

	// routing
	authService := service.NewAuthService(userRepo, tokenService, paymentService, otpService, googleOauthConfig)
	authController := controller.NewAuthController(authService)
	RouteAuth(router, authController, userController)

	RouteUser(router, userController, authMiddlware)
	RoutePet(router, petController, authMiddlware)
	RouteFood(router, foodController, authMiddlware)
	RoutePetFoodPlan(router, petFoodPlanController, authMiddlware)

	ocrService := service.NewOcrService(geminiClient)
	ocrController := controller.NewOcrController(ocrService)
	RouteOcr(router, ocrController, authMiddlware)

	webhookController := controller.NewWebhookController(cfg.STRIPE_WEBHOOK_SECRET, webhookService)
	RouteWebhook(router, webhookController)
}
