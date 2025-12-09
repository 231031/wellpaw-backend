package model

// everytime - PetDetail adjust (add new updated) - PetFoodPlanDetail adjust too
type PetFoodPlanDetail struct {
	ID                uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	FoodPetFoodPlanID uint    `json:"food_pet_food_plan_id"`
	PetDetailID       uint    `json:"pet_detail_id"`
	Amount            float64 `json:"amount"`
	EnergyIntake      float64 `json:"energy_intake"`
	ProteinIntake     float64 `json:"protein_intake"`
	FatIntake         float64 `json:"fat_intake"`

	// Relationships
	FoodPetFoodPlan *FoodPetFoodPlan `gorm:"foreignKey:FoodPetFoodPlanID" json:"food_pet_food_plan,omitempty"`
	// PetDetail       *PetDetail       `gorm:"foreignKey:PetDetailID" json:"pet_detail,omitempty"`
}
