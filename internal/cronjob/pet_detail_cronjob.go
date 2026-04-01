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

func shouldNotifyUpdatePetDetail(p model.Pet) (bool, bool, bool) {
	weightChanged, bcsChanged, alChanged := restrictedChangedInWindow(p.PetDetails)
	return !weightChanged, !bcsChanged, !alChanged
}

func restrictedChangedInWindow(details []model.PetDetail) (bool, bool, bool) {
	bcsChnanged := false
	alChanged := false

	if len(details) == 0 {
		return false, false, false
	}

	oneMonthAgo := time.Now().AddDate(0, -1, 0)
	latestRec := details[0]
	if oneMonthAgo.Before(latestRec.CreatedAt) {
		bcsChnanged = true
		alChanged = true
	}

	lastRec := details[len(details)-1]
	if oneMonthAgo.Before(lastRec.CreatedAt) {
		return true, bcsChnanged, alChanged
	} else if len(details) == 1 {
		// one month ago after or equal to last record
		return false, bcsChnanged, alChanged
	}

	for i := 0; i < len(details)-1; i++ {
		curr := details[i]
		prev := details[i+1]

		if math.Abs(curr.Weight-prev.Weight) > weightComparisonTolerance {
			return true, bcsChnanged, alChanged
		}
	}

	return false, bcsChnanged, alChanged
}

func buildUpdatePetDetailNotification(p model.Pet, updateType model.PetUpdateType) (model.SendNotificationParams, bool) {
	if p.User == nil {
		applogger.LogError(fmt.Sprintf("pet %d has no associated user", p.ID), petLog)
		return model.SendNotificationParams{}, false
	}

	token := strings.TrimSpace(p.User.DeviceToken)
	if token == "" {
		applogger.LogError(fmt.Sprintf("user %d with pet %d has no associated device token", p.User.ID, p.ID), petLog)
		return model.SendNotificationParams{}, false
	}

	petName := strings.TrimSpace(p.Name)
	if petName == "" {
		petName = fmt.Sprintf("Pet %d", p.ID)
	}

	var title string
	var body string
	var updateFlag string

	switch updateType {
	case model.WEIGHT_UPDATE:
		title = fmt.Sprintf("%s ควรอัพเดตน้ำหนัก", petName)
		body = fmt.Sprintf("%s ควรอัพเดตน้ำหนักทุกเดือน", petName)
		updateFlag = "weight_update"
	case model.BCS_UPDATE:
		title = fmt.Sprintf("%s ควรอัพเดต bcs", petName)
		body = fmt.Sprintf("%s ควรอัพเดต bcs ทุกเดือน", petName)
		updateFlag = "bcs_update"
	case model.AL_UPDATE:
		title = fmt.Sprintf("%s ควรอัพเดตระดับกิจกรรม", petName)
		body = fmt.Sprintf("%s ควรอัพเดตระดับกิจกรรมทุกเดือน", petName)
		updateFlag = "activity_update"
	}

	return model.SendNotificationParams{
		Token: token,
		Title: title,
		Body:  body,
		Data: map[string]string{
			"type":   updateFlag,
			"pet_id": strconv.Itoa(int(p.ID)),
		},
	}, true
}

func (cj *mainCronjob) sendUpdatePetDetailNotification(ctx context.Context, notificationsMsg []model.SendNotificationParams) {
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
	notificationsMsg := make([]model.SendNotificationParams, 0)
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
			notiWeight, notiBcs, notiAl := shouldNotifyUpdatePetDetail(p)
			if notiWeight {
				noti, valid := buildUpdatePetDetailNotification(p, model.WEIGHT_UPDATE)
				fmt.Printf("weight : %+v\n", noti)
				if valid {
					notificationsMsg = append(notificationsMsg, noti)
				}
			}
			if notiBcs {
				noti, valid := buildUpdatePetDetailNotification(p, model.BCS_UPDATE)
				fmt.Printf("bcs : %+v\n", noti)

				if valid {
					notificationsMsg = append(notificationsMsg, noti)
				}
			}
			if notiAl {
				noti, valid := buildUpdatePetDetailNotification(p, model.AL_UPDATE)
				fmt.Printf("al : %+v\n", noti)

				if valid {
					notificationsMsg = append(notificationsMsg, noti)
				}
			}
		}

		lastID = pets[len(pets)-1].ID
	}

	if len(notificationsMsg) == 0 {
		applogger.LogInfo("no pets require monthly pet detail reminder", petLog)
		return
	}

	cj.sendUpdatePetDetailNotification(ctx, notificationsMsg)
}
