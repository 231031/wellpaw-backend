package cronjob

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/231031/wellpaw-backend/internal/applogger"
	"github.com/231031/wellpaw-backend/internal/model"
	"gorm.io/gorm"
)

const weightComparisonTolerance = 0.000001

func shouldNotifyUpdatePetDetail(p model.Pet) bool {
	return !restrictedChangedInWindow(p.PetDetails)
}

func restrictedChangedInWindow(details []model.PetDetail) bool {
	if len(details) == 0 {
		return false
	}

	oneMonthAgo := time.Now().AddDate(0, -1, 0)
	if len(details) == 1 {
		if oneMonthAgo.Before(details[0].CreatedAt) {
			return true
		}
		return false
	}

	for i := 0; i < len(details)-1; i++ {
		curr := details[i]
		prev := details[i+1]

		if math.Abs(curr.Weight-prev.Weight) > weightComparisonTolerance {
			return true
		}
	}

	return false
}

func buildUpdatePetDetailNotification(p model.Pet) (model.SendNotificationParams, bool) {
	if p.User == nil {
		return model.SendNotificationParams{}, false
	}

	token := strings.TrimSpace(p.User.DeviceToken)
	if token == "" {
		return model.SendNotificationParams{}, false
	}

	petName := strings.TrimSpace(p.Name)
	if petName == "" {
		petName = fmt.Sprintf("Pet %d", p.ID)
	}

	return model.SendNotificationParams{
		Token: token,
		Title: fmt.Sprintf("%s should update weight and bcs", petName),
		Body:  fmt.Sprintf("%s should update detail every one month", petName),
		Data: map[string]string{
			"type":   "pet_detail_update",
			"pet_id": strconv.Itoa(int(p.ID)),
		},
	}, true
}

func (cj *mainCronjob) sendUpdatePetDetailNotification(ctx context.Context, pets []model.Pet) {
	notificationsMsg := make([]model.SendNotificationParams, 0, len(pets))
	for _, p := range pets {
		notificationMsg, ok := buildUpdatePetDetailNotification(p)
		if !ok {
			if p.User == nil {
				applogger.LogError(fmt.Sprintf("skip pet detail notification for pet_id=%d: user not loaded", p.ID), petLog)
				continue
			}

			applogger.LogError(fmt.Sprintf("skip pet detail notification for pet_id=%d user_id=%d: empty device token", p.ID, p.User.ID), petLog)
			continue
		}

		fmt.Printf("updated %d : %+v\n", p.ID, notificationMsg)
		notificationsMsg = append(notificationsMsg, notificationMsg)
	}

	if len(notificationsMsg) == 0 {
		applogger.LogInfo("no valid pet detail notifications to send", petLog)
		return
	}

	resp, err := cj.fcmService.SendNotifications(ctx, notificationsMsg)
	if err != nil {
		applogger.LogError(fmt.Sprintf("failed to send pet detail notifications (%d) : %v", len(notificationsMsg), err), petLog)
		return
	}
	if resp == nil {
		applogger.LogInfo("fcm response is nil for pet detail notifications", petLog)
		return
	}

	for i, r := range resp.Responses {
		if !r.Success {
			token := ""
			if i < len(notificationsMsg) {
				token = notificationsMsg[i].Token
			}
			errMsg := fmt.Sprintf("token failed: %s : %v", token, r.Error)
			applogger.LogError(errMsg, petLog)
		}
	}
}

func (cj *mainCronjob) NotificateUpdatePetDetail() {
	ctx, cancel := context.WithTimeout(context.Background(), cj.defaultTimeout)
	defer cancel()

	var lastID uint
	var petsNeedUpdate []model.Pet
	for {
		pets, err := cj.petRepo.GetPetsLatestOneMonthPetDetail(ctx, lastID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				break
			}

			applogger.LogError("failed to get pets to check pet detail", petLog)
			return
		}

		if len(pets) == 0 {
			break
		}

		for _, p := range pets {
			if shouldNotifyUpdatePetDetail(p) {
				petsNeedUpdate = append(petsNeedUpdate, p)
			}
		}

		lastID = pets[len(pets)-1].ID
	}

	if len(petsNeedUpdate) == 0 {
		applogger.LogInfo("no pets require monthly pet detail reminder", petLog)
		return
	}

	cj.sendUpdatePetDetailNotification(ctx, petsNeedUpdate)
}
