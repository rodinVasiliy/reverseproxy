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

		r.Name = dto.Name
		r.Enabled = dto.Enabled
		r.Actions = make([]primitive.ObjectID, 0, len(dto.Actions))
		for _, action := range dto.Actions {
			id, err := primitive.ObjectIDFromHex(action)
			if err != nil {
				log.Printf("failed to parse id: %v", err)
				httpx.BadRequest(c, "invalid action id")
				return
			}
			r.Actions = append(r.Actions, id)
		}

		for i, policyId := range dto.Overrides {
			pId, err := primitive.ObjectIDFromHex(policyId.ID)
			if err != nil {
				log.Printf("failed to parse id: %v", err)
				httpx.BadRequest(c, "invalid policy ID")
				return
			}
			actionIDs := make([]primitive.ObjectID, 0, len(dto.Overrides[i].Actions))
			for _, action := range dto.Overrides[i].Actions {
				id, err := primitive.ObjectIDFromHex(action)
				if err != nil {
					log.Printf("failed to parse id: %v", err)
					httpx.BadRequest(c, "invalid action id")
					return
				}
				actionIDs = append(actionIDs, id)
			}
			err = as.AddOverrideToPolicy(c, pId, r.ID, actionIDs)
			if err != nil {
				log.Printf("failed to add override to policy: %v", err)
				httpx.InternalError(c)
				return
			}
		}

		err = s.Update(c, r)
		if err != nil {
			log.Printf("failed to update rule: %v", err)
			httpx.InternalError(c)
			return
		}

		c.Status(204)
	}
}
