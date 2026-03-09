package cronjob

import (
	"context"
	"errors"
	"fmt"

	"github.com/231031/wellpaw-backend/internal/applogger"
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

			applogger.LogError("failed to get active plans to daily update food amount", logAt)
			return
		}

		if len(calendars) == 0 {
			applogger.LogInfo("calendars don't have notification", logAt)
			break
		}

		for _, c := range calendars {
			fmt.Printf("cal : %+v\n", c)
		}

		lastID = calendars[len(calendars)-1].ID
	}
}
