package model

import "time"

type PetDetail struct {
	ID                 uint          `gorm:"primaryKey;autoIncrement" json:"id"`
	PetID              uint          `json:"pet_id"`
	Weight             float64       `json:"weight"`
	ActivityLevel      ActivityLevel `json:"activity_level"`
	BCS                BcsType       `json:"bcs"` // Body Condition Score
	IsAdult            bool          `json:"is_adult"`
	Lactation          bool          `json:"lactation"`
	Gestation          bool          `json:"gestation"`
	GestationStartDate time.Time     `gorm:"type:date" json:"gestation_startdate"`
	Neutered           bool          `json:"neutered"`
	Energy             float64       `json:"energy"`
	Protein            float64       `json:"protein"` // Fixed typo from 'protien'
	Fat                float64       `json:"fat"`
	CreatedAt          time.Time     `json:"created_at"`

	// Relationships
	// Pet                *Pet                 `gorm:"foreignKey:PetID" json:"pet,omitempty"`
	PetFoodPlanDetails []PetFoodPlanDetail  `gorm:"foreignKey:PetDetailID" json:"pet_food_plan_details,omitempty"`
	PetFoodPlanHistory []PetFoodPlanHistory `gorm:"foreignKey:PetDetailID" json:"pet_food_plan_history,omitempty"`
}
