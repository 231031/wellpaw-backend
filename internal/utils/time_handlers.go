package utils

import (
	"fmt"
	"time"
)

func ConvertStripeTimeToTimeStr(stripeTime int64) time.Time {
	thaiLocation, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		fmt.Println("Error loading timezone:", err)
		return time.Time{}
	}

	timeLocThai := time.Unix(stripeTime, 0).In(thaiLocation)

	return timeLocThai
}
