package dto

import (
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Ограничиваем название файла определенным шаблоном, чтобы не было непонятных символов
var safeFileName = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
var upstreamRegexp = regexp.MustCompile(`^(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}|localhost):\d{2,5}$`)
var hostnameRegex = regexp.MustCompile(`^([a-zA-Z0-9-]{1,63}\.)+[a-zA-Z]{2,}$`)

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

func webAppNameValidator(fl validator.FieldLevel) bool {
	v := fl.Field().String()

	if !safeFileName.MatchString(v) {
		return false
	}

	return true
}

func upstreamValidator(fl validator.FieldLevel) bool {
	v := fl.Field().String()

	return upstreamRegexp.MatchString(v)
}

func hostnameValidator(fl validator.FieldLevel) bool {
	v := fl.Field().String()

	if v == "" {
		return false
	}

	// если в качестве хоста используется wildcard
	if strings.HasPrefix(v, "*.") {
		return isValidHostname(strings.Trim(v, "*."))
	}
	return isValidHostname(v)
}

func isValidHostname(h string) bool {
	if len(h) > 253 { // по стандарту rfc больше нельзя
		return false
	}
	return hostnameRegex.MatchString(h)
}

var Validate *validator.Validate

func init() {
	Validate = validator.New()

	Validate.RegisterValidation("certfilename", certFilenameValidator)
	Validate.RegisterValidation("keyfilename", keyFilenameValidator)
	Validate.RegisterValidation("webappname", webAppNameValidator)
	Validate.RegisterValidation("upstream", upstreamValidator)
	Validate.RegisterValidation("host", hostnameValidator)
}
