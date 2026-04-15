package handler

import (
	"log"
	action "reverseproxy/internal/domain/action"
	dto "reverseproxy/internal/dto"
	"reverseproxy/internal/httpx"

	"github.com/gin-gonic/gin"
)

func RegisterActionRoutes(r *gin.RouterGroup, s *action.Service) {
	g := r.Group("/actions")
	{
		g.GET("", getActions(s))
	}
}

func getActions(s *action.Service) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		actions, err := s.FindAll(ctx.Request.Context())
		if err != nil {
			log.Printf("failed to find all actions in get actions api method %v", err)
			httpx.InternalError(ctx)
			return
		}
		dtos := make([]dto.ActionDTO, 0, len(actions))
		for _, a := range actions {
			dtos = append(dtos, *dto.ActionToDTO(&a))
		}
		ctx.JSON(200, dtos)
	}
}
