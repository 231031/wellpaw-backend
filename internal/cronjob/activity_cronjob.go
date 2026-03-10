package cronjob

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/231031/wellpaw-backend/internal/applogger"
	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/231031/wellpaw-backend/internal/utils"
	"gorm.io/gorm"
)

func (cj *mainCronjob) NotificateActivity() {
	logAt := "activity cronjob"
	ctx, cancel := context.WithTimeout(context.Background(), cj.defaultTimeout)
	defer cancel()

	var lastID uint
	for {
		calendars, err := cj.petCalendarRepo.GetActiveActivityCalendar(ctx, lastID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				break
			}

			applogger.LogError("failed to get current activity calendars", logAt)
			return
		}

		if len(calendars) == 0 {
			applogger.LogInfo("calendars don't have notification", logAt)
			break
		}

		notificationsMsg := make([]model.SendNotificationParams, 0)
		for _, c := range calendars {
			if len(c.ActivityEvents) == 0 {
				continue
			}

			frequently := ""
			if c.Frequently != nil {
				frequently = c.Frequently.String()
			}

			calendarName := strings.TrimSpace(c.Name)
			if calendarName == "" {
				calendarName = "Pet activity reminder"
			}

			token := strings.TrimSpace(c.ActivityEvents[0].Pet.User.DeviceToken)
			if token == "" {
				applogger.LogError(fmt.Sprintf("empty device token for user_id=%d", c.ActivityEvents[0].Pet.User.ID), logAt)
				continue
			}

			var bodyStart string
			if *c.Frequently == model.NOT {
				bodyStart = c.Type.String()
			} else {
				bodyStart = fmt.Sprintf("%s %s", c.Frequently.String(), c.Type.String())
			}

			bodyStr := "Reminder for "
			for idx, event := range c.ActivityEvents {
				petName := strings.TrimSpace(event.Pet.Name)
				if petName == "" {
					petName = fmt.Sprintf("pet_%d", event.PetID)
				}

				if idx == len(c.ActivityEvents)-1 {
					bodyStr += petName
				} else {
					bodyStr = bodyStr + petName + ","
				}
			}

			notificationMsg := model.SendNotificationParams{
				Token: token,
				Title: calendarName,
				Body:  bodyStart + " " + bodyStr,
				Data: map[string]string{
					"start_datetime": utils.ConvertTimeToThaiTimezone(c.StartDatetime).String(),
					"frequently":     frequently,
				},
			}
			notificationsMsg = append(notificationsMsg, notificationMsg)
		}

		if len(notificationsMsg) == 0 {
			applogger.LogInfo(fmt.Sprintf("no valid notifications in batch (last_id=%d)", lastID), logAt)
			lastID = calendars[len(calendars)-1].ID
			break
		}

		resp, err := cj.fcmService.SendNotifications(ctx, notificationsMsg)
		if err != nil {
			applogger.LogError(fmt.Sprintf("failed to send notifications (%d) : %v", len(notificationsMsg), err), logAt)
			lastID = calendars[len(calendars)-1].ID
		}
		if resp == nil {
			applogger.LogInfo("fcm response is nil", logAt)
			lastID = calendars[len(calendars)-1].ID
		}

		for i, r := range resp.Responses {
			if !r.Success {
				token := ""
				if i < len(notificationsMsg) {
					token = notificationsMsg[i].Token
				}
				errMsg := fmt.Sprintf("token failed: %s : %v", token, r.Error)
				applogger.LogError(errMsg, logAt)
			}
		}
		lastID = calendars[len(calendars)-1].ID
	}
}
