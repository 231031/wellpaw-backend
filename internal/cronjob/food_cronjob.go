package cronjob

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/231031/wellpaw-backend/internal/applogger"
	"github.com/231031/wellpaw-backend/internal/model"
	"gorm.io/gorm"
)

func (cj *mainCronjob) UpdateQuatityFoodDaily() {
	logAt := "food cronjob"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var lastID uint
	for {
		p, err := cj.petFoodPlanRepo.GetNextActiveFoodPlanByID(ctx, lastID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				break
			}

			applogger.LogError("failed to get active plans to daily update food amount", logAt)
			return
		}
		lastID = p.ID

		var fqInPlans []model.FoodQuantity
		insufficinetAmount := false
		for _, fp := range p.FoodPetFoodPlans {
			if len(fp.PetFoodPlanDetails) == 0 {
				applogger.LogInfo(fmt.Sprintf("food pet food plan : id (%d) not have pet food plan detials\n", fp.ID), logAt)
				insufficinetAmount = true
				break
			}

			amountIntake := fp.PetFoodPlanDetails[0].Amount
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
		if insufficinetAmount {
			err = cj.petFoodPlanRepo.InactivePetFoodPlan(ctx, []model.PetFoodPlan{{
				ID:     p.ID,
				Active: false,
			}})
			if err != nil {
				msg := fmt.Sprintf("failed to daily inactive plan %d that has insufficient food amount: %v", p.ID, err)
				applogger.LogError(msg, logAt)
			}
		} else {
			err = cj.foodRepo.UpdateDailyFoodAmount(ctx, fqInPlans)
			if err != nil {
				msg := fmt.Sprintf("failed to daily update amount in plan %d : %v", p.ID, err)
				applogger.LogError(msg, logAt)
			}
		}
	}

	// notification to notify user

}
