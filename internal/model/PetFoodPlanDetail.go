package model

import "time"

type PetFoodPlanDetail struct {
	ID                 uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	FoodPetFoodPlanID  uint      `gorm:"not null" json:"food_pet_food_plan_id"`
	PetFoodPlanTotalID uint      `gorm:"column:pet_food_plan_totals_id;not null" json:"pet_food_plan_totals_id"`
	Amount             float64   `gorm:"not null" json:"amount"`
	EnergyIntake       float64   `gorm:"not null" json:"energy_intake"`
	ProteinIntake      float64   `gorm:"not null" json:"protein_intake"`
	FatIntake          float64   `gorm:"not null" json:"fat_intake"`
	CreatedAt          time.Time `gorm:"not null" json:"created_at"`

	// Relationships
	FoodPetFoodPlan  *FoodPetFoodPlan  `gorm:"foreignKey:FoodPetFoodPlanID" json:"food_pet_food_plan,omitempty"`
	PetFoodPlanTotal *PetFoodPlanTotal `gorm:"foreignKey:PetFoodPlanTotalID;references:ID" json:"pet_food_plan_total,omitempty"`
}
