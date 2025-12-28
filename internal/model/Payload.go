package model

type LoginPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginGooglePayload struct {
	AuthCode    string `json:"auth_code"`
	DeviceToken string `json:"device_token"`
}

type FoodDetailResponse struct {
	Energy  float64 `json:"energy"`
	Protein float64 `json:"protein"`
	Fat     float64 `json:"fat"`
	Moist   float64 `json:"moist"`
}
