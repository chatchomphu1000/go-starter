package http

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// AppValidator wraps go-playground/validator for use with Echo.
type AppValidator struct {
	validator *validator.Validate
}

// NewAppValidator creates a new AppValidator instance.
func NewAppValidator() *AppValidator {
	v := validator.New()
	return &AppValidator{validator: v}
}

// Validate validates the given struct using go-playground/validator tags.
func (av *AppValidator) Validate(i interface{}) error {
	if err := av.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
	}
	return nil
}
