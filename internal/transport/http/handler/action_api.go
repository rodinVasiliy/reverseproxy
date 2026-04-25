package handler

import (
	"log"
	"reverseproxy/internal/domain/action"
	dtoAction "reverseproxy/internal/dto/action"
	"reverseproxy/internal/httpx"
	"reverseproxy/internal/mapper"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func RegisterActionRoutes(r *gin.RouterGroup, s *action.Service) {
	g := r.Group("/actions")
	{
		g.GET("", getActions(s))
		g.GET("/:id", getAction(s))
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
		actionDTOS := make([]dtoAction.ActionResponse, 0, len(actions))
		for _, a := range actions {
			actionDTOS = append(actionDTOS, *mapper.ToActionResponse(a))
		}
		ctx.JSON(200, actionDTOS)
	}
}

func getAction(s *action.Service) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := primitive.ObjectIDFromHex(ctx.Param("id"))
		if err != nil {
			log.Printf("failed to parse id %v", err)
			httpx.InternalError(ctx)
			return
		}
		actionDoc, err := s.FindById(ctx.Request.Context(), id)
		if err != nil {
			log.Printf("failed to find action %v", err)
			httpx.InternalError(ctx)
			return
		}
		response := mapper.ToActionResponse(*actionDoc)
		ctx.JSON(200, response)
	}
}
