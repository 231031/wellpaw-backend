package model

// everytime - PetDetail adjust (add new updated) - PetFoodPlanDetail adjust too
type PetFoodPlanDetail struct {
	ID                uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	FoodPetFoodPlanID uint    `gorm:"not null" json:"food_pet_food_plan_id"`
	PetDetailID       uint    `gorm:"not null" json:"pet_detail_id"`
	Amount            float64 `gorm:"not null" json:"amount"`
	EnergyIntake      float64 `gorm:"not null" json:"energy_intake"`
	ProteinIntake     float64 `gorm:"not null" json:"protein_intake"`
	FatIntake         float64 `gorm:"not null" json:"fat_intake"`

	// Relationships
	FoodPetFoodPlan *FoodPetFoodPlan `gorm:"foreignKey:FoodPetFoodPlanID" json:"food_pet_food_plan,omitempty"`
	// PetDetail       *PetDetail       `gorm:"foreignKey:PetDetailID" json:"pet_detail,omitempty"`
}
