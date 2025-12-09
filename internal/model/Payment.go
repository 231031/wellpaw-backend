package model

import "time"

type Payment struct {
	ID                    uint              `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID                uint              `gorm:"not null" json:"user_id"`
	Price                 int               `gorm:"not null" json:"price"`
	StripePaymentID       string            `gorm:"type:varchar(255)" json:"stripe_payment_id"`
	Status                PaymentStatusType `gorm:"default:0;not null" json:"status"`
	Currency              string            `gorm:"type:varchar(255)" json:"currency"`
	SubscriptionPeriodEnd time.Time         `gorm:"not null" json:"subscription_period_end"`
	CreatedAt             time.Time         `gorm:"not null" json:"created_at"`
	UpdatedAt             time.Time         `gorm:"not null" json:"updated_at"`

	// Relationships
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
