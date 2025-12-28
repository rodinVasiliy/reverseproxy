package httpx

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

func validationMessage(field, tag string) string {
	switch tag {
	case "required":
		return fmt.Sprintf("field %s is required", field)
	case "min":
		return fmt.Sprintf("field %s is too short", field)
	case "max":
		return fmt.Sprintf("field %s is too long", field)
	case "certfilename":
		return "certificate file must be .pem, .crt or .cer"
	case "keyfilename":
		return "key file must be .pem or .key"
	default:
		return "invalid value"
	}
}

func validationErrorsToMap(err validator.ValidationErrors) map[string]string {
	out := make(map[string]string)

	for _, fe := range err {
		field := fe.Field() // CertFileName
		tag := fe.Tag()     // certfilename

		out[field] = validationMessage(field, tag)
	}

	return out
}
