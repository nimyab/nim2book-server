package register

type Input struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type Output struct {
	Success bool `json:"success"`
}
