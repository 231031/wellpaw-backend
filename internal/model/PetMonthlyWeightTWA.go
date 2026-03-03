package model

type AvgPercentWeightChangePerMonth struct {
	StartPercent      float64 `json:"start_percent"`
	EndPercent        float64 `json:"end_percent"`
	CalculatedPercent float64 `json:"calculated_percent"`
	TotalMonthsUsed   int     `json:"total_months_used"`
	Message           string  `json:"message,omitempty"`
}
