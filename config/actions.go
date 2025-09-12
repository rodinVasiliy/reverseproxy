package config

import (
	"net/http"
)

type Action struct {
	name   string
	action func(http.ResponseWriter, *http.Request)
}
