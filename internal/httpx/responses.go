package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, APIError{
		Code:    "bad_request",
		Message: message,
	})
}

func NotFound(c *gin.Context, resource string) {
	c.JSON(http.StatusNotFound, APIError{
		Code:    "not_found",
		Message: resource + " not found",
	})
}

func InternalError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, APIError{
		Code:    "internal_server_error",
		Message: "Internal server error",
	})
}

func ValidationError(c *gin.Context, ve validator.ValidationErrors) {
	c.JSON(http.StatusBadRequest, gin.H{
		"code":   "validation_error",
		"fields": validationErrorsToMap(ve),
	})
}
