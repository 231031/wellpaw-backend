package cronjob

import (
	"fmt"
	"time"

	"github.com/231031/wellpaw-backend/internal/repository"
	"github.com/231031/wellpaw-backend/internal/service"
	"github.com/robfig/cron/v3"
)

var (
	foodLog     = "food cronjob"
	activityLog = "activity cronjob"
	petLog      = "pet cronjob"
)

type mainCronjob struct {
	calculationSerivce service.CalculationService
	fcmService         service.FcmService
	petRepo            repository.PetRepository
	foodRepo           repository.FoodRepository
	petFoodPlanRepo    repository.PetFoodPlanRepository
	petCalendarRepo    repository.PetCalendarRepository
	defaultTimeout     time.Duration
}

func CreateCronjob(calSerivce service.CalculationService, fcmService service.FcmService, petRepo repository.PetRepository, foodRepo repository.FoodRepository, petFoodPlanRepo repository.PetFoodPlanRepository, petCalendarRepo repository.PetCalendarRepository) {
	main := &mainCronjob{
		calculationSerivce: calSerivce,
		fcmService:         fcmService,
		petRepo:            petRepo,
		foodRepo:           foodRepo,
		petFoodPlanRepo:    petFoodPlanRepo,
		petCalendarRepo:    petCalendarRepo,
		defaultTimeout:     10 * time.Second,
	}

	location, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		fmt.Printf("Error loading location: %v\n", err)
		return
	}

	c := cron.New(cron.WithLocation(location))
	c.AddFunc("59 23 * * *", main.UpdateQuatityFoodDaily)
	c.AddFunc("00 09 * * *", main.NotificateUpdatePetDetail)
	c.AddFunc("*/5 * * * *", main.NotificateActivity)
	c.Start()
}
