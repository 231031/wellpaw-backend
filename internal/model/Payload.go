package model

type LoginPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginGooglePayload struct {
	AuthCode    string `json:"auth_code"`
	DeviceToken string `json:"device_token"`
}

type PetFoodAnalysisResponse struct {
	Energy   *float64 `json:"energy,omitempty"`
	Protein  *float64 `json:"protein,omitempty"`
	Fat      *float64 `json:"fat,omitempty"`
	Moisture *float64 `json:"moisture,omitempty"`
}
