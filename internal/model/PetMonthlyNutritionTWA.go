package model

type PetMonthlyNutritionTWA struct {
	Month              uint    `json:"month"`
	TotalEnergyIntake  float64 `json:"total_energy_intake"`
	TotalProteinIntake float64 `json:"total_protein_intake"`
	TotalFatIntake     float64 `json:"total_fat_intake"`
	Energy             float64 `json:"energy"`
	Protein            float64 `json:"protein"`
	Fat                float64 `json:"fat"`
	Weight             float64 `json:"weight"`
	ActivityLevel      int     `json:"activity_level"`
}
