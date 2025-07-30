package validator

import "github.com/go-playground/validator/v10"

type Validator struct {
	validator *validator.Validate
}

func New() *Validator {
	return &Validator{
		validator: validator.New(),
	}
}

func (ev *Validator) Validate(i interface{}) error {
	return ev.validator.Struct(i)
}
