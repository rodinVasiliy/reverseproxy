package handler

import (
	"log"
	"net/http"
	"reverseproxy/internal/domain/policy"
	"reverseproxy/internal/dto"
	policy2 "reverseproxy/internal/dto/policy"
	"reverseproxy/internal/httpx"
	"reverseproxy/internal/mapper"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func RegisterPolicyRoutes(r *gin.RouterGroup, s *policy.Service) {
	g := r.Group("/policies")
	{
		g.GET("", getPolicies(s))
		g.GET("/:id", getPolicy(s))
		g.DELETE("/:id", deletePolicy(s))
	}
}

func getPolicies(s *policy.Service) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		items, err := s.List(ctx.Request.Context())
		if err != nil {
			log.Printf("failed to find all policies in get policy api: %v", err)
			httpx.InternalError(ctx)
			return
		}

		responses := mapper.ToPolicyListItems(items)
		ctx.JSON(200, responses)
	}
}

func deletePolicy(s *policy.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			log.Printf("failed to parse id in delete policy api: %v", err)
			httpx.BadRequest(c, "invalid id")
			return
		}

		p, err := s.FindById(c, id)
		if err != nil {
			log.Printf("failed to find policy in delete policy api: %v", err)
			httpx.InternalError(c)
			return
		}

		webappNames, err := s.GetWebappsByPolicyId(c, id)
		if err != nil {
			log.Printf("failed to find webapp names by policy id in delete policy api: %v", err)
			httpx.InternalError(c)
			return
		}
		if len(webappNames) > 0 {
			log.Printf("policy %v in use in webapps: %v", p.Name, webappNames)

			c.JSON(http.StatusBadRequest, gin.H{
				"code":    "policy_in_use",
				"message": "Policy is used",
				"webapps": webappNames,
			})
			return
		}

		err = s.Delete(c.Request.Context(), p)
		if err != nil {
			log.Printf("failed to delete policy in delete policy api: %v", err)
			httpx.InternalError(c)
			return
		}

		c.Status(204)
	}
}

func getPolicy(s *policy.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			log.Printf("failed to parse id in get policy api: %v", err)
			httpx.BadRequest(c, "invalid id")
			return
		}

		p, err := s.GetPolicyDetailById(c, id)
		if err != nil {
			log.Printf("failed to find policy in get policy api: %v", err)
			httpx.InternalError(c)
			return
		}

		response := mapper.ToPolicyDetail(p)
		c.JSON(200, response)
	}
}

func updatePolicy(s *policy.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			log.Printf("failed to parse id in update policy api: %v", err)
			httpx.BadRequest(c, "invalid id")
			return
		}

		var policyDTO policy2.Dto
		if err := c.ShouldBindJSON(&policyDTO); err != nil {
			httpx.BadRequest(c, "invalid json body")
			return
		}

		if err := dto.Validate.Struct(&policyDTO); err != nil {
			handleValidationError(err, c)
			return
		}

		p := &policy.Policy{
			ID:   id,
			Name: policyDTO.Name,
			WL:   policyDTO.WL,
		}

		err = s.Update(c, p)
		if err != nil {
			log.Printf("failed to update policy in update policy api: %v", err)
			httpx.InternalError(c)
			return
		}
		c.Status(204)
	}

}
