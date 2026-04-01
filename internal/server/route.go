package server

import (
	"firebase.google.com/go/v4/messaging"
	"github.com/231031/wellpaw-backend/internal/controller"
	"github.com/231031/wellpaw-backend/internal/cronjob"
	"github.com/231031/wellpaw-backend/internal/middleware"
	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/231031/wellpaw-backend/internal/repository"
	"github.com/231031/wellpaw-backend/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/google/generative-ai-go/genai"
	"github.com/redis/go-redis/v9"
	"github.com/stripe/stripe-go/v84"
	"gorm.io/gorm"
)

func RouteAuth(router fiber.Router, authController controller.AuthController, authMiddleware middleware.AuthMiddleware) {
	authRoute := router.Group("/auth")
	authRoute.Post("/register", authController.CreateUser)
	authRoute.Post("/login", authController.LoginUser)
	authRoute.Post("/login/google", authController.LoginUserWithGoogle)
	authRoute.Post("/refreshtoken", authController.RefreshToken)
	authRoute.Post("/logout", authMiddleware.AuthorizeUser(), authController.LogoutUser)
	authRoute.Post("/otp", authController.RequestOTP)
	authRoute.Post("/resetpassword", authController.ResetPassword)
}

func RouteUser(router fiber.Router, userController controller.UserController, authMiddleware middleware.AuthMiddleware) {
	userRoute := router.Group("/user", authMiddleware.AuthorizeUser())
	userRoute.Patch("/devicetoken", userController.UpdateDeviceToken)
	userRoute.Get("/", userController.GetUserAllInfo)

	userRoute.Get("/notification/food", userController.ManageFoodNotification)
	userRoute.Get("/notification/calendar", userController.ManageCalendarNotification)
	userRoute.Get("/notification/pet", userController.ManageUpdatePetNotification)

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
	petRoute.Get("/:pet_id", petController.GetPetByPetID)
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
	foodRoute.Get("/:food_id", foodController.GetFoodByIDAndUserID)
	foodRoute.Post("/quantity", foodController.CreateNewFoodQuantity)
	foodRoute.Patch("/", foodController.UpdateFoodDetail)
	foodRoute.Delete("/:food_id", foodController.SoftDeleteFood)
}

func RoutePetFoodPlan(router fiber.Router, petFoodPlanController controller.PetFoodPlanController, authMiddleware middleware.AuthMiddleware) {
	petFoodPlanRoute := router.Group("/foodplan", authMiddleware.AuthorizeUser())
	petFoodPlanRoute.Post("/calculate", petFoodPlanController.CalculatePetFoodPlan)
	petFoodPlanRoute.Post("/", petFoodPlanController.CreatePetFoodPlan)
	// petFoodPlanRoute.Put("/amount", petFoodPlanController.UpdateFeedingAmountFromUser)
	petFoodPlanRoute.Get("/:pet_id", petFoodPlanController.GetLastestActivePlanDetailByPet)
}

func RoutePetCalendar(router fiber.Router, petCalendarController controller.PetCalendarController, authMiddleware middleware.AuthMiddleware) {
	router.Get("/calendars", authMiddleware.AuthorizeUser(), petCalendarController.GetPetCalendarsByUserID)
	router.Get("/calendars/sum", authMiddleware.AuthorizeUser(), petCalendarController.GetCurrentMonthCalendarTypeSummaryByUserID)
	router.Get("/calendars/:pet_id", authMiddleware.AuthorizeUser(), petCalendarController.GetPetCalendarsByPetID)

	petCalendarRoute := router.Group("/calendar", authMiddleware.AuthorizeUser())
	petCalendarRoute.Post("/", petCalendarController.CreatePetCalendar)
}

func RouteWebhook(router fiber.Router, webhookController controller.WebhookController) {
	webhookRoute := router.Group("/webhook")
	webhookRoute.Post("/subscription", webhookController.HandleAllWebhook)
}

func RouteOcr(router fiber.Router, ocrController controller.OcrController, authMiddleware middleware.AuthMiddleware) {
	ocrRoute := router.Group("/ocr", authMiddleware.AuthorizeUser())
	ocrRoute.Post("/request", ocrController.ProcessOcrRequest)
}

func RouteDisease(router fiber.Router, diseaseController controller.DiseaseController, authMiddleware middleware.AuthMiddleware) {
	router.Get("/diseases", authMiddleware.AuthorizeUser(), diseaseController.GetPetSkinImagesByUserID)
	router.Get("/diseases/:pet_id", authMiddleware.AuthorizeUser(), diseaseController.GetPetSkinImagesByPetID)

	diseaseRoute := router.Group("/disease", authMiddleware.AuthorizeUser())
	diseaseRoute.Post("/predict", diseaseController.PredictDisease)
	diseaseRoute.Post("/predict/unknown", authMiddleware.AuthorizeUser(), diseaseController.PredictDiseaseUnknown)
	diseaseRoute.Patch("/labeled", diseaseController.LabeledPetSkinDisease)
}

func CreateRoute(router fiber.Router, db *gorm.DB, redisClient *redis.Client, geminiClient *genai.Client, stripeClient *stripe.Client, firebaseStorage *model.FirebaseStorage, fcmClient *messaging.Client, cfg *Cfg) {
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
	userService := service.NewUserService(userRepo, paymentService)
	userController := controller.NewUserController(userService)

	energyReqService := service.NewEnergyRequirementService()
	nutritientReqService := service.NewNutritientRequirementService()
	expectedWeightService := service.NewExpectedWeightService()
	calculationService := service.NewCalculationService(energyReqService, nutritientReqService, expectedWeightService)

	freeTierRepo := repository.NewFreeTierUsageRepository(db, redisClient)
	freeTierValidationService := service.NewFreeTierUsageValidationService(userRepo, freeTierRepo)

	foodRepo := repository.NewFoodRepository(db)
	foodService := service.NewFoodService(calculationService, foodRepo, freeTierValidationService)
	foodController := controller.NewFoodController(foodService)

	petRepo := repository.NewPetRepository(db)
	petFoodPlanRepo := repository.NewPetFoodPlanRepository(db)
	petCalendarRepo := repository.NewPetCalendarRepository(db)

	petFoodPlanService := service.NewPetFoodPlanService(calculationService, petFoodPlanRepo, petRepo, foodRepo, freeTierValidationService)
	petFoodPlanController := controller.NewPetFoodPlanController(petFoodPlanService)

	petService := service.NewPetService(calculationService, petRepo, petFoodPlanRepo, freeTierValidationService)
	petController := controller.NewPetController(petService)

	petCalendarService := service.NewPetCalendarService(petCalendarRepo)
	petCalendarController := controller.NewPetCalendarController(petCalendarService)

	petSkinImageRepo := repository.NewPetSkinImageRepository(db)
	modelService := service.NewModelService(cfg.MODEL_BASE_API)
	diseaseService := service.NewDiseaseService(modelService, petRepo, petSkinImageRepo, freeTierValidationService, firebaseStorage)
	diseaseController := controller.NewDiseaseController(diseaseService)

	// routing
	authService := service.NewAuthService(userRepo, tokenService, paymentService, otpService, googleOauthConfig)
	authController := controller.NewAuthController(authService)
	RouteAuth(router, authController, authMiddlware)

	RouteUser(router, userController, authMiddlware)
	RoutePet(router, petController, authMiddlware)
	RouteFood(router, foodController, authMiddlware)
	RoutePetFoodPlan(router, petFoodPlanController, authMiddlware)
	RoutePetCalendar(router, petCalendarController, authMiddlware)
	RouteDisease(router, diseaseController, authMiddlware)

	ocrService := service.NewOcrService(geminiClient, freeTierValidationService)
	ocrController := controller.NewOcrController(ocrService)
	RouteOcr(router, ocrController, authMiddlware)

	fcmService := service.NewFCMService(fcmClient)

	// webhookService := service.NewWebhookService(userRepo, fcmService)
	// webhookController := controller.NewWebhookController(cfg.STRIPE_WEBHOOK_SECRET, webhookService)
	// RouteWebhook(router, webhookController)

	// cronjob
	cronjob.CreateCronjob(calculationService, fcmService, petRepo, foodRepo, petFoodPlanRepo, petCalendarRepo)
}
