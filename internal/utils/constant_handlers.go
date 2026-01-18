package utils

import "github.com/231031/wellpaw-backend/internal/model"

func ConvertIntervalToTierType(interval string) model.TierType {
	switch interval {
	case "week":
		return model.WEEKS
	case "month":
		return model.MONTHS
	case "year":
		return model.YEARS
	default:
		return model.FREE
	}
}
