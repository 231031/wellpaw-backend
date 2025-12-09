package model

type Food struct {
	ID        uint     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint     `json:"user_id"`
	ImagePath string   `gorm:"type:varchar(512)" json:"image_path"`
	Name      string   `gorm:"type:varchar(255)" json:"name"`
	Brand     string   `gorm:"type:varchar(255)" json:"brand"`
	Type      FoodType `json:"type"`
	Quantity  int      `json:"quantity"`
	Weight    float64  `json:"weight"`
	Energy    float64  `json:"energy"`
	Protein   float64  `json:"protein"`
	Fat       float64  `json:"fat"`
	Moist     float64  `json:"moist"`

	// Relationships
	User            *User             `gorm:"foreignKey:UserID" json:"user,omitempty"`
	FoodPetFoodPlan []FoodPetFoodPlan `gorm:"foreignKey:FoodID" json:"food_pet_food_plans,omitempty"`
}
