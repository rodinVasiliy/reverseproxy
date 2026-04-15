package handler

import (
	"log"
	policy "reverseproxy/internal/domain/policy"
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
		responses, err := s.List(ctx.Request.Context())
		if err != nil {
			log.Printf("failed to find all policies in get policy api: %v", err)
			httpx.InternalError(ctx)
			return
		}
		ctx.JSON(200, responses)
	}
}
