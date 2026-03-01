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
}

func CreateCronjob(calSerivce service.CalculationService, foodRepo repository.FoodRepository, petFoodPlanRepo repository.PetFoodPlanRepository) {
	main := &mainCronjob{
		calculationSerivce: calSerivce,
		foodRepo:           foodRepo,
		petFoodPlanRepo:    petFoodPlanRepo,
	}

	location, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		fmt.Printf("Error loading location: %v\n", err)
		return
	}

	c := cron.New(cron.WithLocation(location))
	c.AddFunc("59 23 * * *", main.UpdateQuatityFoodDaily)
	c.Start()
}
