package handler

import (
	"errors"
	"log"
	webapp "reverseproxy/internal/domain/webapp"
	dto "reverseproxy/internal/dto"
	webappDto "reverseproxy/internal/dto"
	"reverseproxy/internal/httpx"
	mongorepository "reverseproxy/internal/infrastructure/mongo"
	manager "reverseproxy/internal/waf/proxy"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func RegisterWebAppRoutes(r *gin.RouterGroup, s *webapp.Service, manager *manager.Manager) {
	g := r.Group("/webapps")
	{
		g.GET("", getWebApps(s))
		g.POST("", createWebApp(s, manager))
		g.PUT("/:id", updateWebApp(s, manager))
		g.DELETE("/:id", deleteWebApp(s, manager))
		g.GET("/:id", getWebApp(s))
	}
}

// Здесь используется webapp.WebappResponce в качестве выхода
func getWebApps(s *webapp.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		apps, err := s.List(c.Request.Context())
		if err != nil {
			log.Printf("failed to find all web apps: %v", err)
			httpx.InternalError(c)
			return
		}

		c.JSON(200, apps)
	}
}

func getWebApp(s *webapp.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			log.Printf("failed to parse id: %v", err)
			httpx.BadRequest(c, "invalid id")
		}

		w, err := s.FindById(c.Request.Context(), id)
		if err != nil {
			log.Printf("failed to find web app: %v", err)
			httpx.InternalError(c)
		}

		wDto := dto.WebAppToDTO(*w)
		c.JSON(200, wDto)
	}

}

func createWebApp(s *webapp.Service, manager *manager.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var webAppDTO webappDto.WebAppDTO
		if err := c.ShouldBindJSON(&webAppDTO); err != nil {
			log.Printf("invalid json:%s", webAppDTO.String())
			httpx.BadRequest(c, "invalid json")
			return
		}
		log.Printf("Create web app request: %s\n", webAppDTO.String())

		if err := dto.Validate.Struct(&webAppDTO); err != nil {
			var ve validator.ValidationErrors
			if errors.As(err, &ve) {
				log.Println("Validation error!!!")
				for _, fe := range ve {
					log.Printf("validation filed[%s] failed: [%s]", fe.Field(), fe.Tag())
				}
				httpx.ValidationError(c, ve)
				return
			}
			log.Println("invalid request")
			httpx.BadRequest(c, "invalid request")
			return
		}

		webapp, err := webappDto.DTOToWebApp(webAppDTO)
		if err != nil {
			log.Printf("failed to convert dto to webapp")
			httpx.BadRequest(c, err.Error())
			return
		}

		id, err := s.Insert(c.Request.Context(), *webapp)
		if err != nil {
			log.Printf("failed to insert web app to db: %v", err)
			httpx.InternalError(c)
		}

		err = manager.SetProxyToManager(webapp)
		if err != nil {
			log.Printf("failed to add proxy to manager: %s", err)
			httpx.InternalError(c)
		}

		c.JSON(201, gin.H{"id": id.Hex()})
	}
}

func updateWebApp(s *webapp.Service, manager *manager.Manager) gin.HandlerFunc {
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

		err = manager.SetProxyToManager(webapp)
		if err != nil {
			log.Printf("failed to add proxy to manager: %s", err)
			httpx.InternalError(c)
		}

		c.Status(204)
	}
}

func deleteWebApp(s *webapp.Service, manager *manager.Manager) gin.HandlerFunc {
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

		manager.DeleteProxyFromManager(webapp)

		c.Status(204)
	}
}
