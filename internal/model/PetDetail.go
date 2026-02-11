package model

import "time"

type PetDetail struct {
	ID                 uint          `gorm:"primaryKey;autoIncrement" json:"id"`
	PetID              uint          `gorm:"not null" json:"pet_id"`
	Weight             float64       `gorm:"not null" json:"weight"`
	ExpectedWeight     float64       `gorm:"not null" json:"expected_weight"`
	ActivityLevel      ActivityLevel `gorm:"not null" json:"activity_level"`
	AgeRange           AgeType       `gorm:"not null" json:"age_range"`
	BCS                BcsType       `gorm:"not null" json:"bcs"` // Body Condition Score
	Lactation          bool          `gorm:"default:false;not null" json:"lactation"`
	Gestation          bool          `gorm:"default:false;not null" json:"gestation"`
	GestationStartDate time.Time     `gorm:"type:date" json:"gestation_startdate"`
	Neutered           bool          `gorm:"default:false;not null" json:"neutered"`
	Energy             float64       `gorm:"not null" json:"energy"`
	Protein            float64       `gorm:"not null" json:"protein"`
	Fat                float64       `gorm:"not null" json:"fat"`
	CreatedAt          time.Time     `gorm:"not null" json:"created_at"`

	// Relationships
	// Pet                *Pet                 `gorm:"foreignKey:PetID" json:"pet,omitempty"`
	PetFoodPlanDetails []PetFoodPlanDetail  `gorm:"foreignKey:PetDetailID" json:"pet_food_plan_details,omitempty"`
	PetFoodPlanHistory []PetFoodPlanHistory `gorm:"foreignKey:PetDetailID" json:"pet_food_plan_history,omitempty"`
}
