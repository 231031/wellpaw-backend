package cronjob

import (
	"fmt"
	"time"

	"github.com/231031/wellpaw-backend/internal/repository"
	"github.com/231031/wellpaw-backend/internal/service"
	"github.com/robfig/cron/v3"
)

type MainCronjob interface{}

type mainCronjob struct {
	calculationSerivce service.CalculationService
	foodRepo           repository.FoodRepository
	petFoodPlanRepo    repository.PetFoodPlanRepository
	petCalendarRepo    repository.PetCalendarRepository
	defaultTimeout     time.Duration
}

func CreateCronjob(calSerivce service.CalculationService, foodRepo repository.FoodRepository, petFoodPlanRepo repository.PetFoodPlanRepository, petCalendarRepo repository.PetCalendarRepository) {
	main := &mainCronjob{
		calculationSerivce: calSerivce,
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
	c.AddFunc("*/5 * * * *", main.NotificateActivity)
	c.Start()
}
