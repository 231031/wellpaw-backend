package model

type PetMonthlyNutritionTWA struct {
	Year                uint    `json:"year"`
	Month               uint    `json:"month"`
	TotalEnergyIntake   float64 `json:"total_energy_intake"`
	TotalProteinIntake  float64 `json:"total_protein_intake"`
	TotalFatIntake      float64 `json:"total_fat_intake"`
	Energy              float64 `json:"energy"`
	Protein             float64 `json:"protein"`
	Fat                 float64 `json:"fat"`
	Weight              float64 `json:"weight"`
	PercentWeightChange float64 `json:"percent_weight_change"`
	ActivityLevel       int     `json:"activity_level"`
}
