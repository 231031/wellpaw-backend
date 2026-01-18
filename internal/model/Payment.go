package model

import "time"

type Payment struct {
	ID             uint              `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID         uint              `gorm:"not null" json:"user_id"`
	PriceID        string            `gorm:"not null" json:"price_id"`
	SubscriptionID string            `gorm:"type:varchar(255)" json:"subscription_id"`
	Status         PaymentStatusType `gorm:"default:0;not null" json:"status"`
	CreatedAt      time.Time         `gorm:"not null" json:"created_at"`
	UpdatedAt      time.Time         `gorm:"not null" json:"updated_at"`
	// SubscriptionPeriodEnd time.Time         `gorm:"not null" json:"subscription_period_end"`

	// Relationships
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
