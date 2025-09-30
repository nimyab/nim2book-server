package delete_fcm_token

type Input struct {
	FcmToken string `query:"token" validate:"required"`
}

type Output struct {
	Success bool `json:"success"`
}
