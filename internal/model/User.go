package model

import "time"

type User struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Email         string    `gorm:"type:varchar(255)" json:"email"`
	Password      string    `gorm:"type:varchar(255)" json:"password"`
	NotiFood      bool      `json:"noti_food"`
	NotiCalendars bool      `json:"noti_calendars"`
	ProfileFree   int       `json:"profile_free"`
	FoodFree      int       `json:"food_free"`
	FoodPlanFree  int       `json:"food_plan_free"`
	MainFree      int       `json:"main_free"`
	PaymentPlan   TierType  `json:"payment_plan"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// Relationships
	Pets     []Pet     `gorm:"foreignKey:UserID" json:"pets,omitempty"`
	Foods    []Food    `gorm:"foreignKey:UserID" json:"foods,omitempty"`
	Payments []Payment `gorm:"foreignKey:UserID" json:"payments,omitempty"`
}
