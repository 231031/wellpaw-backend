package model

import "time"

type Payment struct {
	ID                    uint              `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID                uint              `json:"user_id"`
	Price                 int               `json:"price"`
	StripePaymentID       string            `gorm:"type:varchar(255)" json:"stripe_payment_id"`
	Status                PaymentStatusType `json:"status"`
	Currency              string            `gorm:"type:varchar(255)" json:"currency"`
	SubscriptionPeriodEnd time.Time         `json:"subscription_period_end"`
	CreatedAt             time.Time         `json:"created_at"`
	UpdatedAt             time.Time         `json:"updated_at"`

	// Relationships
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
