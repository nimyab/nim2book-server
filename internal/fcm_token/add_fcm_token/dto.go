package add_fcm_token

type Input struct {
	FcmToken string `json:"fcmToken" validate:"required"`
}

type Output struct {
	Success bool `json:"success"`
}
