package model

type CupFoodPet struct {
	ID                uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	FoodPetFoodPlanID uint    `gorm:"not null" json:"food_pet_food_plan_id"`
	Grams             float64 `gorm:"not null" json:"unit_type"`
}
