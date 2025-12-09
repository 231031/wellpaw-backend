package model

type Food struct {
	ID        uint     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint     `gorm:"not null" json:"user_id"`
	ImagePath string   `gorm:"type:varchar(512)" json:"image_path"`
	Name      string   `gorm:"not null;type:varchar(255)" json:"name"`
	Brand     string   `gorm:"not null;type:varchar(255)" json:"brand"`
	Type      FoodType `gorm:"not null" json:"type"`
	Quantity  int      `gorm:"not null" json:"quantity"`
	Weight    float64  `gorm:"not null" json:"weight"`
	Energy    float64  `gorm:"not null" json:"energy"`
	Protein   float64  `gorm:"not null" json:"protein"`
	Fat       float64  `gorm:"not null" json:"fat"`
	Moist     float64  `gorm:"not null" json:"moist"`

	// Relationships
	User            *User             `gorm:"foreignKey:UserID" json:"user,omitempty"`
	FoodPetFoodPlan []FoodPetFoodPlan `gorm:"foreignKey:FoodID" json:"food_pet_food_plans,omitempty"`
}
