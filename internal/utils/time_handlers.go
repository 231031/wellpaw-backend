package utils

import (
	"time"
)

const thaiTimezone = "Asia/Bangkok"

func ConvertTimeToThaiTimezone(dateTime time.Time) time.Time {
	if dateTime.IsZero() {
		return dateTime
	}

	thaiLocation, err := time.LoadLocation(thaiTimezone)
	if err != nil {
		return dateTime
	}

	return dateTime.In(thaiLocation)
}

func ConvertDateToThaiMidnight(date time.Time) time.Time {
	if date.IsZero() {
		return date
	}

	thaiLocation, err := time.LoadLocation(thaiTimezone)
	if err != nil {
		return date
	}

	year, month, day := date.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, thaiLocation)
}

func ConvertStripeTimeToTimeStr(stripeTime int64) time.Time {
	return ConvertTimeToThaiTimezone(time.Unix(stripeTime, 0))
}

func GetMonthRangeInThai(now time.Time) (time.Time, time.Time) {
	thaiNow := ConvertTimeToThaiTimezone(now)
	monthStart := time.Date(thaiNow.Year(), thaiNow.Month(), 1, 0, 0, 0, 0, thaiNow.Location())
	nextMonthStart := monthStart.AddDate(0, 1, 0)

	return monthStart, nextMonthStart
}
