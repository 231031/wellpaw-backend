package model

import "time"

type User struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Email         string    `gorm:"unique;not null;type:varchar(255)" json:"email"`
	Password      string    `gorm:"not null;type:varchar(255)" json:"password,omitempty"`
	DeviceToken   string    `gorm:"not null;type:varchar(512)" json:"device_token"`
	FisrtName     string    `gorm:"not null;type:varchar(255)" json:"first_name"`
	LastName      string    `gorm:"not null;type:varchar(255)" json:"last_name"`
	NotiFood      bool      `gorm:"default:true;not null" json:"noti_food"`
	NotiCalendars bool      `gorm:"default:true;not null" json:"noti_calendars"`
	ProfileFree   int       `gorm:"default:0;not null" json:"profile_free"`
	FoodFree      int       `gorm:"default:0;not null" json:"food_free"`
	FoodPlanFree  int       `gorm:"default:0;not null" json:"food_plan_free"`
	BcsFree       int       `gorm:"default:0;not null" json:"bcs_free"`
	DiseaseFree   int       `gorm:"default:0;not null" json:"disease_free"`
	PaymentPlan   TierType  `gorm:"default:0;not null" json:"payment_plan"`
	CreatedAt     time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt     time.Time `gorm:"not null" json:"updated_at"`

	// Relationships
	Pets     []Pet     `gorm:"foreignKey:UserID" json:"pets,omitempty"`
	Foods    []Food    `gorm:"foreignKey:UserID" json:"foods,omitempty"`
	Payments []Payment `gorm:"foreignKey:UserID" json:"payments,omitempty"`
}
