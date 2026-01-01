package handler

import (
	"errors"
	"log"
	dto "reverseproxy/internal/dto"
	webappDto "reverseproxy/internal/dto"
	"reverseproxy/internal/httpx"
	webapp "reverseproxy/internal/model/webapp"
	mongorepository "reverseproxy/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func RegisterWebAppRoutes(r *gin.RouterGroup, s *webapp.Service) {
	g := r.Group("/webapps")
	{
		g.GET("", getWebApss(s))
		g.POST("", createWebApp(s))
		g.PUT("/:id", updateWebApp(s))
		g.DELETE("/:id", deleteWebApp(s))
	}
}

func getWebApss(s *webapp.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		apps, err := s.FindAll(c.Request.Context())
		if err != nil {
			log.Printf("failed to find all web apps: %v", err)
			httpx.InternalError(c)
			return
		}

		dtos := make([]webappDto.WebAppDTO, 0, len(apps))
		for _, app := range apps {
			dtos = append(dtos, *webappDto.WebAppToDTO(app))
		}

		c.JSON(200, dtos)
	}
}

func createWebApp(s *webapp.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var webAppDTO webappDto.WebAppDTO
		if err := c.ShouldBindJSON(&webAppDTO); err != nil {
			httpx.BadRequest(c, "invalid json")
			return
		}

		if err := dto.Validate.Struct(&webAppDTO); err != nil {
			var ve validator.ValidationErrors
			if errors.As(err, &ve) {
				httpx.ValidationError(c, ve)
				return
			}
			httpx.BadRequest(c, "invalid request")
			return
		}

		webapp, err := webappDto.DTOToWebApp(webAppDTO)
		if err != nil {
			httpx.BadRequest(c, err.Error())
			return
		}

		id, err := s.Insert(c.Request.Context(), *webapp)
		if err != nil {
			log.Printf("failed to insert web app to db: %v", err)
			httpx.InternalError(c)
		}

		c.JSON(201, gin.H{"id": id.Hex()})
	}
}

func updateWebApp(s *webapp.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			httpx.BadRequest(c, "invalid id")
			return
		}

		var dto webappDto.WebAppDTO
		if err := c.ShouldBindJSON(&dto); err != nil {
			httpx.BadRequest(c, "invalid json body")
			return
		}

		webapp, err := webappDto.DTOToWebApp(dto)
		webapp.ID = id
		if err != nil {
			httpx.BadRequest(c, err.Error())
			return
		}

		if err := s.Edit(c.Request.Context(), webapp); err != nil {
			log.Printf("failed to edit web app: %v", err)
			httpx.InternalError(c)
			return
		}

		c.Status(204)
	}
}

func deleteWebApp(s *webapp.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			httpx.BadRequest(c, "invalid id")
			return
		}
		webapp, err := s.FindById(c.Request.Context(), id)
		if err != nil {
			if errors.Is(err, mongorepository.ErrNotFound) {
				httpx.NotFound(c, "web app "+webapp.Name)
				return
			}
			httpx.InternalError(c)
			return
		}
		if err = s.Delete(c.Request.Context(), webapp); err != nil {
			log.Printf("failed to delete web app %v: %v", webapp.Name, err)
			httpx.InternalError(c)
			return
		}
		c.Status(204)
	}
}
