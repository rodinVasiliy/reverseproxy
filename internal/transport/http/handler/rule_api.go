package handler

import (
	"log"
	"net/http"
	appRule "reverseproxy/internal/app/rule"
	"reverseproxy/internal/domain/rule"
	"reverseproxy/internal/dto"
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
		g.POST("", createRule(as))
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

		var rDto ruleDto.RuleDto
		err = c.ShouldBindJSON(&rDto)
		if err != nil {
			log.Printf("failed to parse rule: %v", err)
			httpx.BadRequest(c, "invalid id")
			return
		}

		// Валидация Rule
		err = dto.Validate.Struct(&rDto)
		if err != nil {
			handleValidationError(err, c)
			return
		}

		r, err := s.FindById(c.Request.Context(), id)
		if err != nil {
			log.Printf("failed to fetch rule: %v", err)
			httpx.InternalError(c)
			return
		}

		// Обновляем правило, тут же происходит и рекомпиляция правила и связанных с ним политик
		err = as.UpdateRule(c, r, &rDto)
		if err != nil {
			log.Printf("failed to update rule: %v", err)
			httpx.InternalError(c)
			return
		}

		c.Status(204)
	}
}

func createRule(as *appRule.AppRuleService) func(c *gin.Context) {
	return func(c *gin.Context) {
		var ruleDto ruleDto.RuleDto
		err := c.ShouldBindJSON(&ruleDto)
		if err != nil {
			log.Printf("failed to parse rule: %v", err)
			httpx.BadRequest(c, "invalid id")
			return
		}

		// Валидация Rule
		err = dto.Validate.Struct(&ruleDto)
		if err != nil {
			handleValidationError(err, c)
			return
		}

	}
}
