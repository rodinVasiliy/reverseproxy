package handler

import (
	"log"
	"net/http"
	"reverseproxy/internal/domain/rule"
	"reverseproxy/internal/httpx"

	"github.com/gin-gonic/gin"
)

func RegisterRuleRoutes(r *gin.RouterGroup, s *rule.Service) {
	g := r.Group("/rules")
	{
		g.GET("", getRules(s))
	}
}

func getRules(s *rule.Service) func(c *gin.Context) {
	return func(c *gin.Context) {
		rules, err := s.FindAll(c.Request.Context())
		if err != nil {
			log.Printf("failed to fetch rules: %v", err)
			httpx.InternalError(c)
			return
		}

		c.JSON(http.StatusOK, rules)

	}
}
