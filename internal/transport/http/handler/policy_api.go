package handler

import (
	"errors"
	"log"
	"net/http"
	appPolicy "reverseproxy/internal/app/policy"
	"reverseproxy/internal/domain/policy"
	"reverseproxy/internal/dto"
	policyDto "reverseproxy/internal/dto/policy"
	"reverseproxy/internal/httpx"
	"reverseproxy/internal/mapper"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func RegisterPolicyRoutes(r *gin.RouterGroup, s *policy.Service, aps *appPolicy.AppPolicyService) {
	g := r.Group("/policies")
	{
		g.GET("", getPolicies(aps))
		g.GET("/:id", getPolicy(aps))
		g.DELETE("/:id", deletePolicy(s, aps))
		g.PUT("/:id", updatePolicy(s, aps))
		g.POST("", createPolicy(aps))
	}
}

func getPolicies(aps *appPolicy.AppPolicyService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		items, err := aps.List(ctx.Request.Context())
		if err != nil {
			log.Printf("failed to find all policies in get policy api: %v", err)
			httpx.InternalError(ctx)
			return
		}

		responses := mapper.ToPolicyListItems(items)
		ctx.JSON(200, responses)
	}
}

func deletePolicy(s *policy.Service, aps *appPolicy.AppPolicyService) gin.HandlerFunc {
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

		err = aps.CanDeletePolicy(c.Request.Context(), id)
		if err != nil {
			var inUse *policy.PolicyInUseError
			if errors.As(err, &inUse) {
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    "policy_in_use",
					"message": "Policy is used",
					"webapps": inUse.Webapps,
				})
				return
			}

			log.Printf("failed to delete policy in delete policy api: %v", err)
			httpx.InternalError(c)
			return
		}

		err = aps.Delete(c, p, s)
		if err != nil {
			log.Printf("failed to delete policy in delete policy api: %v", err)
			httpx.InternalError(c)
			return
		}

		c.Status(204)
	}
}

func getPolicy(aps *appPolicy.AppPolicyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			log.Printf("failed to parse id in get policy api: %v", err)
			httpx.BadRequest(c, "invalid id")
			return
		}

		p, err := aps.GetPolicyDetailById(c, id)
		if err != nil {
			log.Printf("failed to find policy in get policy api: %v", err)
			httpx.InternalError(c)
			return
		}

		response := mapper.ToPolicyDetail(p)
		c.JSON(200, response)
	}
}

func updatePolicy(s *policy.Service, as *appPolicy.AppPolicyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			log.Printf("failed to parse id in update policy api: %v", err)
			httpx.BadRequest(c, "invalid id")
			return
		}

		var policyDTO policyDto.Dto
		if err := c.ShouldBindJSON(&policyDTO); err != nil {
			httpx.BadRequest(c, "invalid json body")
			return
		}
		if policyDTO.WL == nil {
			policyDTO.WL = []string{}
		}

		if err := dto.Validate.Struct(&policyDTO); err != nil {
			handleValidationError(err, c)
			return
		}

		p, err := s.FindById(c, id)
		if err != nil {
			log.Printf("failed to find policy by id: %v", err)
			httpx.InternalError(c)
			return
		}
		p.Name = policyDTO.Name
		p.WL = policyDTO.WL

		err = as.Update(c, s, p)
		if err != nil {
			log.Printf("failed to update policy: %v", err)
			httpx.InternalError(c)
			return
		}

		c.Status(204)
	}

}

func createPolicy(as *appPolicy.AppPolicyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var policyDTO policyDto.Dto
		if err := c.ShouldBindJSON(&policyDTO); err != nil {
			httpx.BadRequest(c, "invalid json body")
			return
		}
		if policyDTO.WL == nil {
			policyDTO.WL = []string{}
		}

		if err := dto.Validate.Struct(&policyDTO); err != nil {
			handleValidationError(err, c)
			return
		}

		err := as.Create(c, policyDTO)
		if err != nil {
			log.Printf("failed to create policy: %v", err)
			httpx.InternalError(c)
			return
		}

		c.Status(201)
	}
}
