package cronjob

import (
	"context"
	"fmt"
	"time"

	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/231031/wellpaw-backend/internal/repository"
	"github.com/231031/wellpaw-backend/internal/service"
	"github.com/robfig/cron/v3"
)

type MainCronjob interface {
	UpdateQuatityFoodDaily()
	checkQuatityFoodDaily(p *model.PetFoodPlan) ([]model.FoodQuantity, bool)
	checkNextQuatityFoodDaily(ctx context.Context, lastID uint) (*model.PetFoodPlan, bool)
	sendFoodQuantitiesNotification(ctx context.Context, planInfficient []model.NotificationPlan)

	NotificateActivity()
	getCalendarMessageBody(c model.PetCalendar) string
	sendCalendarNotification(ctx context.Context, notificationsMsg []model.SendNotificationParams)
}

var (
	foodLog     = "food cronjob"
	activityLog = "activity cronjob"
)

type mainCronjob struct {
	calculationSerivce service.CalculationService
	fcmService         service.FcmService
	foodRepo           repository.FoodRepository
	petFoodPlanRepo    repository.PetFoodPlanRepository
	petCalendarRepo    repository.PetCalendarRepository
	defaultTimeout     time.Duration
}

func CreateCronjob(calSerivce service.CalculationService, fcmService service.FcmService, foodRepo repository.FoodRepository, petFoodPlanRepo repository.PetFoodPlanRepository, petCalendarRepo repository.PetCalendarRepository) {
	main := &mainCronjob{
		calculationSerivce: calSerivce,
		fcmService:         fcmService,
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
