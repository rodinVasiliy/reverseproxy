package dto

import (
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ограничиваем название файла определенным шаблоном, чтобы не было непонятных символов
var safeFileName = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

func certFilenameValidator(fl validator.FieldLevel) bool {
	v := fl.Field().String()

	if v == "" || strings.Contains(v, "/") || strings.Contains(v, "\\") || strings.Contains(v, "..") {
		return false
	}

	if !safeFileName.MatchString(v) {
		return false
	}

	return strings.HasSuffix(v, ".crt") || strings.HasSuffix(v, ".pem") || strings.HasSuffix(v, ".cer")
}

func keyFilenameValidator(fl validator.FieldLevel) bool {
	v := fl.Field().String()

	if v == "" || strings.Contains(v, "/") || strings.Contains(v, "\\") || strings.Contains(v, "..") {
		return false
	}

	if !safeFileName.MatchString(v) {
		return false
	}

	return strings.HasSuffix(v, ".pem") || strings.HasSuffix(v, ".key")
}

var Validate *validator.Validate

func init() {
	Validate = validator.New()

	Validate.RegisterValidation("certfilename", certFilenameValidator)
	Validate.RegisterValidation("keyfilename", keyFilenameValidator)
}
