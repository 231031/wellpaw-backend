package model

import "time"

type Pet struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint      `json:"user_id"`
	ImagePath string    `gorm:"type:varchar(512)" json:"image_path"`
	Name      string    `gorm:"type:varchar(255)" json:"name"`
	Type      int       `json:"type"`
	Breed     string    `gorm:"type:varchar(255)" json:"breed"`
	SexType   SexType   `json:"sex_type"`
	BirthDate time.Time `gorm:"type:date" json:"birth_date"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	// User           *User                 `gorm:"foreignKey:UserID" json:"user,omitempty"`
	PetDetails     []PetDetail           `gorm:"foreignKey:PetID" json:"pet_details,omitempty"`
	PetFoodPlans   []PetFoodPlan         `gorm:"foreignKey:PetID" json:"pet_food_plans,omitempty"`
	PetSkinImages  []PetSkinImage        `gorm:"foreignKey:PetID" json:"pet_skin_images,omitempty"`
	ActivityEvents []PetActivityCalendar `gorm:"foreignKey:PetID" json:"activity_events,omitempty"`
}
