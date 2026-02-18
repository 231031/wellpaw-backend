package model

type LoginData struct {
	User  *User      `json:"user"`
	Token *TokenPair `json:"token"`
}

type LoginResponse struct {
	Status int        `json:"status"`
	Data   *LoginData `json:"data"`
}

type UserResponse struct {
	Status int   `json:"status"`
	Data   *User `json:"data"`
}

type SubscriptionPlanResponse struct {
	Status int                 `json:"status"`
	Data   []*SubscriptionPlan `json:"data"`
}

type SubscriptionHistoryPagination struct {
	Subscriptions []*SubscriptionHistory `json:"subscriptions"`
	LastID        string                 `json:"last_id"`
}

type SubscriptionHistoryPaginationResponse struct {
	Status int                            `json:"status"`
	Data   *SubscriptionHistoryPagination `json:"data"`
}

type SubscriptionHistoryResponse struct {
	Status int                  `json:"status"`
	Data   *SubscriptionHistory `json:"data"`
}

type PaymentIntentResponse struct {
	Status int             `json:"status"`
	Data   *PaymentInvoice `json:"data"`
}

type PetFoodAnalysisResponse struct {
	Energy   *float64 `json:"energy,omitempty"`
	Protein  *float64 `json:"protein,omitempty"`
	Fat      *float64 `json:"fat,omitempty"`
	Moisture *float64 `json:"moisture,omitempty"`
}

type OcrPetFoodResponse struct {
	Status int                      `json:"status"`
	Data   *PetFoodAnalysisResponse `json:"data"`
}

type PetResponse struct {
	Status int         `json:"status"`
	Data   *PetPayload `json:"data"`
}

type FoodResponse struct {
	Status int   `json:"status"`
	Data   *Food `json:"data"`
}

type PetFoodPlanDetailResponse struct {
	Pet         *Pet         `json:"pet_info,omitempty"`
	PetDetail   *PetDetail   `json:"pet_detail,omitempty"`
	PetFoodPlan *PetFoodPlan `json:"pet_food_plan,omitempty"`
}

type PetFoodPlanResponse struct {
	Status int                        `json:"status"`
	Data   *PetFoodPlanDetailResponse `json:"data"`
}

// type SubscriptionScheduleResponse struct {
// 	Status int                            `json:"status"`
// 	Data   []*stripe.SubscriptionSchedule `json:"data"`
// }
