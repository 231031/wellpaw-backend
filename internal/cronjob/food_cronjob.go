package cronjob

import (
	"context"
	"errors"
	"fmt"

	"github.com/231031/wellpaw-backend/internal/applogger"
	"github.com/231031/wellpaw-backend/internal/model"
	"gorm.io/gorm"
)

// 14, 20,23
func (cj *mainCronjob) sendFoodQuantitiesNotification(ctx context.Context, planInfficient []model.NotificationPlan) {
	notificationsMsg := make([]model.SendNotificationParams, 0)
	for _, pi := range planInfficient {
		p := pi.Plan

		notificationMsg := model.SendNotificationParams{}
		if pi.OutTime == model.NOW {
			notificationMsg = model.SendNotificationParams{
				Token: p.Pet.User.DeviceToken,
				Title: fmt.Sprintf("แผนอาหาร %s ควรอัพเดตปริมาณอาหาร", p.Name),
				Body:  fmt.Sprintf("อาหารในคลังของแผน %s หมดแล้วและ WellPaw จะไม่สามารถติดตามปริมาณอาหารของคุณได้แม้แผนจะยังคงใช้งานอยู่", p.Name),
			}
		} else {
			notificationMsg = model.SendNotificationParams{
				Token: p.Pet.User.DeviceToken,
				Title: fmt.Sprintf("แผนอาหาร %s ควรอัพเดตปริมาณอาหาร", p.Name),
				Body:  fmt.Sprintf("อาหารในคลังของแผน %s ใกล้หมดแล้วและ WellPaw จะไม่สามารถติดตามปริมาณอาหารของคุณได้แม้แผนจะยังคงใช้งานอยู่", p.Name),
			}
		}
		notificationsMsg = append(notificationsMsg, notificationMsg)

	}

	resp, err := cj.fcmService.SendNotifications(ctx, notificationsMsg)
	if err != nil {
		applogger.LogError(fmt.Sprintf("failed to send notifications (%d) : %v", len(notificationsMsg), err), foodLog)
	}
	if resp == nil {
		applogger.LogInfo("fcm response is nil", foodLog)
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

func (cj *mainCronjob) UpdateQuatityFoodDaily() {
	ctx, cancel := context.WithTimeout(context.Background(), cj.defaultTimeout)
	defer cancel()

	var lastID uint
	var planInfficient []model.NotificationPlan
	for {
		p, err := cj.petFoodPlanRepo.GetNextActiveFoodPlanByID(ctx, lastID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				break
			}

			applogger.LogError("failed to get active plans to daily update food amount", foodLog)
			return
		}

		fqInPlans, insufficinetAmount := cj.checkQuatityFoodDaily(p, 1)
		if insufficinetAmount {
			notiPlan := model.NotificationPlan{
				Plan:    *p,
				OutTime: model.NOW,
			}
			planInfficient = append(planInfficient, notiPlan)
			lastID = p.ID
			continue
		} else {
			err = cj.foodRepo.UpdateDailyFoodAmount(ctx, fqInPlans)
			if err != nil {
				msg := fmt.Sprintf("failed to daily update amount in plan %d : %v", p.ID, err)
				applogger.LogError(msg, foodLog)
			}

			// check 3 days
			_, insufficinetAmount := cj.checkQuatityFoodDaily(p, 4)
			if insufficinetAmount {
				notiPlan := model.NotificationPlan{
					Plan:    *p,
					OutTime: model.NEXT,
				}
				planInfficient = append(planInfficient, notiPlan)
			}
		}

		lastID = p.ID
	}

	if len(planInfficient) > 0 {
		cj.sendFoodQuantitiesNotification(ctx, planInfficient)
	}

}

func (cj *mainCronjob) checkQuatityFoodDaily(p *model.PetFoodPlan, rangeDay int) ([]model.FoodQuantity, bool) {
	fqInPlans := []model.FoodQuantity{}
	insufficinetAmount := false
	for _, fp := range p.FoodPetFoodPlans {
		if len(fp.PetFoodPlanDetails) == 0 {
			applogger.LogInfo(fmt.Sprintf("food pet food plan : id (%d) not have pet food plan detials\n", fp.ID), foodLog)
			insufficinetAmount = true
			break
		}

		amountIntake := fp.PetFoodPlanDetails[0].Amount * float64(rangeDay)
		var checkFq []model.FoodQuantity
		for idx, fq := range fp.Food.FoodQuantities {
			foodAmount := fq.Amount
			if amountIntake > foodAmount {
				if idx == len(fp.Food.FoodQuantities)-1 {
					insufficinetAmount = true
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
		if insufficinetAmount {
			break
		}
	}

	return fqInPlans, insufficinetAmount
}
