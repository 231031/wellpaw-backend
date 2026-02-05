package model

import "time"

type LoginPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginGooglePayload struct {
	AuthCode    string `json:"auth_code"`
	DeviceToken string `json:"device_token"`
}

type PaymentMethodUpdatePayload struct {
	PaymentMethodID string `json:"payment_method_id"`
}

type StartSubscriptionPayload struct {
	SubscriptionPlanID string `json:"subscription_plan_id"`
	PaymentMethodID    string `json:"payment_method_id,omitempty"`
}

type SubscriptionPlan struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Features      []string `json:"features"`
	Amount        int64    `json:"amount"`
	Currency      string   `json:"currency"`
	Interval      string   `json:"interval"`
	IntervalCount int      `json:"interval_count"`
}

type PaymentInvoice struct {
	ClientSecret           string `json:"client_secret"`
	PaymentIntentStatus    string `json:"payment_intent_status"`
	SubscriptionStatus     string `json:"subscription_status"`
	DefaultPaymentMethodID string `json:"default_payment_method_id"`
	Amount                 int64  `json:"amount"`
}

type UpdateSubscriptionPayload struct {
	NewSubscriptionPlanID string `json:"new_subscription_plan_id"`
}

type SubscriptionHistory struct {
	SubscriptionID     string    `json:"subscription_id"`
	SubscriptionStatus string    `json:"subscription_status"`
	InvoiceID          string    `json:"invoice_id"`
	InvoiceStatus      string    `json:"invoice_status"`
	PaymentIntentID    string    `json:"payment_intent_id"`
	PriceID            string    `json:"price_id"`
	AmountPaid         int64     `json:"amount_paid"`
	AmountDue          int64     `json:"amount_due"`
	Amount             int64     `json:"amount"`
	PeriodStart        time.Time `json:"period_start"`
	PeriodEnd          time.Time `json:"period_end"`
	Tier               TierType  `json:"tier"`
}

type PetFoodPlanDetailsPayload struct {
	Name   string   `json:"name"`
	PetID  uint     `json:"pet_id"`
	FoodID []uint   `json:"food_id"`
	Unit   UnitType `json:"unit"`
}

type CreatePetPayload struct {
	PetInfo   Pet       `json:"pet_info"`
	PetDetail PetDetail `json:"pet_detail"`
}
