package utils

import (
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
)

var (
	validate *validator.Validate
	once     sync.Once
)

func GetValidator() *validator.Validate {
	once.Do(func() {
		validate = validator.New(validator.WithRequiredStructEnabled())
	})

	return validate
}

func ValidateStruct(s interface{}) error {
	return GetValidator().Struct(s)
}

// GetValidationErrors returns formatted validation errors as a slice of error messages
func GetValidationErrors(err error) []string {
	var errorMessages []string

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := e.Field()
			tag := e.Tag()

			var message string
			switch tag {
			case "required":
				message = field + " is required"
			case "min":
				if e.Kind() == 8 { // string
					message = field + " must be at least " + e.Param() + " characters"
				} else { // number
					message = field + " must be at least " + e.Param()
				}
			case "max":
				if e.Kind() == 8 { // string
					message = field + " must be at most " + e.Param() + " characters"
				} else { // number
					message = field + " must be at most " + e.Param()
				}
			case "gte":
				message = field + " must be greater than or equal to " + e.Param()
			case "lte":
				message = field + " must be less than or equal to " + e.Param()
			case "gt":
				message = field + " must be greater than " + e.Param()
			case "lt":
				message = field + " must be less than " + e.Param()
			case "email":
				message = field + " must be a valid email address"
			default:
				message = field + " is invalid"
			}

			errorMessages = append(errorMessages, message)
		}
	}

	return errorMessages
}

// GetValidationErrorsMap returns formatted validation errors as a map
func GetValidationErrorsMap(err error) map[string]string {
	errorMap := make(map[string]string)

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := e.Field()
			tag := e.Tag()

			var message string
			switch tag {
			case "required":
				message = field + " is required"
			case "min":
				if e.Kind() == 8 { // string
					message = "Must be at least " + e.Param() + " characters"
				} else { // number
					message = "Must be at least " + e.Param()
				}
			case "max":
				if e.Kind() == 8 { // string
					message = "Must be at most " + e.Param() + " characters"
				} else { // number
					message = "Must be at most " + e.Param()
				}
			case "gte":
				message = "Must be greater than or equal to " + e.Param()
			case "lte":
				message = "Must be less than or equal to " + e.Param()
			case "gt":
				message = "Must be greater than " + e.Param()
			case "email":
				message = "Must be a valid email address"
			default:
				message = "Invalid value"
			}

			// Convert field name to lowercase with underscores for user-friendly display
			fieldName := strings.ToLower(strings.ReplaceAll(field, "_", " "))
			errorMap[fieldName] = message
		}
	}

	return errorMap
}
