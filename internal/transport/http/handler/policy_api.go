package handler

import (
	"log"
	policy "reverseproxy/internal/domain/policy"
	dto "reverseproxy/internal/dto"
	"reverseproxy/internal/httpx"

	"github.com/gin-gonic/gin"
)

func RegisterPolicyRoutes(r *gin.RouterGroup, s *policy.Service) {
	g := r.Group("/policies")
	{
		g.GET("", getPolicies(s))
	}
}

func getPolicies(s *policy.Service) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		policies, err := s.FindAll(ctx.Request.Context())
		if err != nil {
			log.Printf("failed to find all policies in get policy api: %v", err)
			httpx.InternalError(ctx)
		}
		dtos := make([]dto.PolicyDTO, 0, len(policies))
		for _, p := range policies {
			dtos = append(dtos, *dto.PolicyToDTO(&p))
		}
		ctx.JSON(200, dtos)
	}
}
