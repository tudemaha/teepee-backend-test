package validator

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

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

// FormatErrors extracts friendly string error messages from a validation error
func FormatErrors(err error) []string {
	var errs []string
	if fieldErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range fieldErrors {
			errs = append(errs, fmt.Sprintf("%s failed validation: %s", e.Field(), e.Tag()))
		}
		return errs
	}
	if err != nil {
		return []string{err.Error()}
	}
	return nil
}
