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

type RequestOTPPayload struct {
	Email string `json:"email"`
}

type ResetPasswordPayload struct {
	Email             string `json:"email"`
	OTP               string `json:"otp"`
	Password          string `json:"password"`
	ConfirmedPassword string `json:"confirmed_password"`
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

type FoodPlanFoodPayload struct {
	FoodID      uint     `json:"food_id"`
	GramsPerCup *float64 `json:"grams_per_cup,omitempty"`
}

type CreatePetFoodPlanPayload struct {
	Name  string                `json:"name"`
	PetID uint                  `json:"pet_id"`
	Unit  UnitType              `json:"unit"`
	Foods []FoodPlanFoodPayload `json:"foods"`
}

type AmountFoodDetail struct {
	FoodPetFoodPlanID uint    `json:"food_pet_food_plan_id"`
	Amount            float64 `json:"amount"`
}
type AdjustAmountFoodInPetFoodPlanPayload struct {
	PetFoodPlanID      uint               `json:"pet_food_plan_id"`
	PetFoodPlanDetails []AmountFoodDetail `json:"pet_food_plan_details"`
}

type PetPayload struct {
	PetInfo   Pet       `json:"pet_info"`
	PetDetail PetDetail `json:"pet_detail"`
}
