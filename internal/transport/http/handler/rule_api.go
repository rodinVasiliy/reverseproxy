package handler

import (
	"log"
	"net/http"
	appRule "reverseproxy/internal/app/rule"
	"reverseproxy/internal/domain/rule"
	ruleDto "reverseproxy/internal/dto/rule"
	"reverseproxy/internal/httpx"
	"reverseproxy/internal/mapper"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func RegisterRuleRoutes(r *gin.RouterGroup, s *rule.Service, as *appRule.AppRuleService) {
	g := r.Group("/rules")
	{
		g.GET("", getRules(s))
		g.GET("/:id", getRule(as))
		g.PUT("/:id", editRule(s, as))
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

		ruleListItems := mapper.ToRuleListItems(rules)
		c.JSON(http.StatusOK, ruleListItems)
	}
}

func getRule(s *appRule.AppRuleService) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			log.Printf("failed to parse id: %v", err)
			httpx.BadRequest(c, "invalid id")
			return
		}

		ruleResponse, err := s.RuleResponse(c, id)
		if err != nil {
			log.Printf("failed to fetch rule: %v", err)
			httpx.InternalError(c)
			return
		}

		c.JSON(http.StatusOK, ruleResponse)
	}

}

func editRule(s *rule.Service, as *appRule.AppRuleService) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			log.Printf("failed to parse id: %v", err)
			httpx.BadRequest(c, "invalid id")
			return
		}

		var dto ruleDto.RuleDto
		err = c.ShouldBindJSON(&dto)
		if err != nil {
			log.Printf("failed to parse rule: %v", err)
			httpx.BadRequest(c, "invalid id")
			return
		}

		r, err := s.FindById(c.Request.Context(), id)
		if err != nil {
			log.Printf("failed to fetch rule: %v", err)
			httpx.InternalError(c)
			return
		}

		err = as.UpdateRule(c, r, &dto)
		if err != nil {
			log.Printf("failed to update rule: %v", err)
			httpx.InternalError(c)
			return
		}

		c.Status(204)
	}
}
