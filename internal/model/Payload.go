package model

import "time"

type LoginPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginGooglePayload struct {
	AuthCode    string `json:"auth_code" validate:"required"`
	DeviceToken string `json:"device_token" validate:"required"`
}

type RequestOTPPayload struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordPayload struct {
	Email             string `json:"email" validate:"required,email"`
	OTP               string `json:"otp" validate:"required"`
	Password          string `json:"password" validate:"required"`
	ConfirmedPassword string `json:"confirmed_password" validate:"required"`
}

type PaymentMethodUpdatePayload struct {
	PaymentMethodID string `json:"payment_method_id" validate:"required"`
}

type StartSubscriptionPayload struct {
	SubscriptionPlanID string `json:"subscription_plan_id" validate:"required"`
	PaymentMethodID    string `json:"payment_method_id,omitempty" validate:"required"`
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
	NewSubscriptionPlanID string `json:"new_subscription_plan_id" validate:"required"`
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
	Name   string    `json:"name" validate:"required"`
	PetID  uint      `json:"pet_id" validate:"required"`
	FoodID []uint    `json:"food_id" validate:"required"`
	Unit   *UnitType `json:"unit" validate:"required,oneof=0 1"`
}

type FoodPlanFoodPayload struct {
	FoodID      uint     `json:"food_id" validate:"required"`
	GramsPerCup *float64 `json:"grams_per_cup,omitempty"`
}

type CreatePetFoodPlanPayload struct {
	Name  string                `json:"name" validate:"required"`
	PetID uint                  `json:"pet_id" validate:"required"`
	Unit  *UnitType             `json:"unit" validate:"required,oneof=0 1"`
	Foods []FoodPlanFoodPayload `json:"foods" validate:"required"`
}

type AmountFoodDetail struct {
	FoodPetFoodPlanID uint    `json:"food_pet_food_plan_id" validate:"required"`
	Amount            float64 `json:"amount" validate:"required,gt=0"`
}
type AdjustAmountFoodInPetFoodPlanPayload struct {
	PetFoodPlanID      uint               `json:"pet_food_plan_id" validate:"required"`
	PetFoodPlanDetails []AmountFoodDetail `json:"pet_food_plan_details" validate:"required"`
}

type PetPayload struct {
	PetInfo   *Pet       `json:"pet_info" validate:"required"`
	PetDetail *PetDetail `json:"pet_detail" validate:"required"`
}
