package model

import "time"

type PetFoodPlanTotal struct {
	ID                 uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	PetFoodPlanID      uint      `gorm:"not null" json:"pet_food_plan_id"`
	PetDetailID        uint      `gorm:"not null" json:"pet_detail_id"`
	TotalEnergyIntake  float64   `gorm:"not null" json:"total_energy_intake"`
	TotalProteinIntake float64   `gorm:"not null" json:"total_protein_intake"`
	TotalFatIntake     float64   `gorm:"not null" json:"total_fat_intake"`
	CreatedAt          time.Time `gorm:"not null" json:"created_at"`

	// Relationships
	// PetFoodPlan      *PetFoodPlan       `gorm:"foreignKey:PetFoodPlanID" json:"pet_food_plan,omitempty"`
	// PetDetail        *PetDetail         `gorm:"foreignKey:PetDetailID" json:"pet_detail,omitempty"`
	PetFoodPlanDetails []PetFoodPlanDetail  `gorm:"foreignKey:PetFoodPlanTotalID;references:ID" json:"pet_food_plan_details,omitempty"`
	PetFoodPlanHistory []PetFoodPlanHistory `gorm:"foreignKey:PetFoodPlanTotalID;references:ID" json:"pet_food_plan_history,omitempty"`
}
