package cronjob

import (
	"context"
	"errors"
	"fmt"
	"maps"

	"github.com/231031/wellpaw-backend/internal/applogger"
	"github.com/231031/wellpaw-backend/internal/model"
	"gorm.io/gorm"
)

// 14, 20,23
func (cj *mainCronjob) sendFoodQuantitiesNotification(ctx context.Context, foodInfficient map[uint]model.Food) {
	notificationsMsg := make([]model.SendNotificationParams, 0)
	for key, food := range foodInfficient {
		notificationMsg := model.SendNotificationParams{
			Token: food.User.DeviceToken,
			Title: "ควรอัพเดตปริมาณอาหาร",
			Body:  fmt.Sprintf("อาหาร %s ในคลังของแผนอาหารใกล้หมดแล้วและ WellPaw จะไม่สามารถติดตามปริมาณอาหารของคุณได้แม้แผนจะยังคงใช้งานอยู่", food.Name),
			Data: map[string]string{
				"type":    "food_quantity_update",
				"food_id": fmt.Sprintf("%d", key),
			},
		}
		notificationsMsg = append(notificationsMsg, notificationMsg)
	}

	resp, err := cj.fcmService.SendNotifications(ctx, notificationsMsg)
	if err != nil {
		applogger.LogError(fmt.Sprintf("failed to send notifications (%d) : %v", len(notificationsMsg), err), foodLog)
	}
	if resp == nil {
		applogger.LogInfo("fcm response is nil", foodLog)
		return
	}

	for i, r := range resp.Responses {
		if !r.Success {
			token := ""
			if i < len(notificationsMsg) {
				token = notificationsMsg[i].Token
			}
			errMsg := fmt.Sprintf("token failed: %s : %v", token, r.Error)
			applogger.LogError(errMsg, foodLog)
		}
	}
}

func (cj *mainCronjob) mapInsufficientFoods(foodInfficient, newFoods map[uint]model.Food) map[uint]model.Food {
	insufficientFoods := foodInfficient
	maps.Copy(insufficientFoods, newFoods)
	return insufficientFoods
}

func (cj *mainCronjob) UpdateQuatityFoodDaily() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in UpdateQuatityFoodDaily", r)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), cj.defaultTimeout)
	defer cancel()

	var lastID uint
	// var planInfficient []model.NotificationPlan
	foodInfficient := map[uint]model.Food{}

	for {
		p, err := cj.petFoodPlanRepo.GetNextActiveFoodPlanByID(ctx, lastID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				break
			}

			applogger.LogError("failed to get active plans to daily update food amount", foodLog)
			return
		}

		fqInPlans, insufficinetFoods, insufficinetAmount := cj.checkQuatityFoodDaily(p, 1)
		if insufficinetAmount {
			foodInfficient = cj.mapInsufficientFoods(foodInfficient, insufficinetFoods)

			lastID = p.ID
			continue
		} else {
			err = cj.foodRepo.UpdateDailyFoodAmount(ctx, fqInPlans)
			if err != nil {
				msg := fmt.Sprintf("failed to daily update amount in plan %d : %v", p.ID, err)
				applogger.LogError(msg, foodLog)
			}

			// check 3 days
			_, insufficinetFoods, insufficinetAmount := cj.checkQuatityFoodDaily(p, 4)
			if insufficinetAmount {
				foodInfficient = cj.mapInsufficientFoods(foodInfficient, insufficinetFoods)
			}
		}

		lastID = p.ID
	}

	if len(foodInfficient) > 0 {
		cj.sendFoodQuantitiesNotification(ctx, foodInfficient)
	}

}

func (cj *mainCronjob) checkQuatityFoodDaily(p *model.PetFoodPlan, rangeDay int) ([]model.FoodQuantity, map[uint]model.Food, bool) {
	fqInPlans := []model.FoodQuantity{}
	insufficinetFoods := map[uint]model.Food{}
	insufficinetAmount := false
	for _, fp := range p.FoodPetFoodPlans {
		if len(fp.PetFoodPlanDetails) == 0 {
			applogger.LogInfo(fmt.Sprintf("food pet food plan : id (%d) not have pet food plan detials\n", fp.ID), foodLog)
			break
		}

		amountIntake := fp.PetFoodPlanDetails[0].Amount * float64(rangeDay)
		var checkFq []model.FoodQuantity
		for idx, fq := range fp.Food.FoodQuantities {
			foodAmount := fq.Amount
			if amountIntake > foodAmount {
				if idx == len(fp.Food.FoodQuantities)-1 {
					insufficinetAmount = true
					insufficinetFoods[fp.FoodID] = *fp.Food
				} else {
					amountIntake = amountIntake - fq.Amount
					checkFq = append(checkFq, model.FoodQuantity{
						ID:     fq.ID,
						Amount: 0,
					})
				}
			} else {
				if len(checkFq) > 0 {
					fqInPlans = append(fqInPlans, checkFq...)
				}
				fqInPlans = append(fqInPlans, model.FoodQuantity{
					ID:     fq.ID,
					Amount: foodAmount - amountIntake,
				})
				break
			}
		}
		// if insufficinetAmount {
		// 	break
		// }
	}

	return fqInPlans, insufficinetFoods, insufficinetAmount
}
