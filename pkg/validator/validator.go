package validator

import "github.com/go-playground/validator/v10"

// CustomValidator wraps the go-playground validator to satisfy the echo.Validator interface
type CustomValidator struct {
	validator *validator.Validate
}

func New() *CustomValidator {
	return &CustomValidator{
		validator: validator.New(),
	}
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}
