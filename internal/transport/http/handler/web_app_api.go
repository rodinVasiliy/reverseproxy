package handler

import (
	"errors"
	"log"
	appWebapp "reverseproxy/internal/app/webapp"
	"reverseproxy/internal/domain/webapp"
	"reverseproxy/internal/dto"
	webappDto "reverseproxy/internal/dto/webapp"
	"reverseproxy/internal/httpx"
	mongorepository "reverseproxy/internal/infrastructure/mongo"
	"reverseproxy/internal/mapper"
	manager "reverseproxy/internal/waf/proxy"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func RegisterWebAppRoutes(r *gin.RouterGroup, s *webapp.Service, app *appWebapp.AppWebappService, manager *manager.Manager) {
	g := r.Group("/webapps")
	{
		g.GET("", getWebApps(app))
		g.POST("", createWebApp(s, manager))
		g.PUT("/:id", updateWebApp(s, manager))
		g.DELETE("/:id", deleteWebApp(s, manager))
		g.GET("/:id", getWebApp(s))
	}
}

func getWebApps(s *appWebapp.AppWebappService) gin.HandlerFunc {
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

		wDto := mapper.WebappToDto(*w)
		c.JSON(200, wDto)
	}

}

func createWebApp(s *webapp.Service, manager *manager.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var webAppDTO webappDto.WebappDto
		if err := c.ShouldBindJSON(&webAppDTO); err != nil {
			log.Printf("invalid json:%s", webAppDTO.String())
			httpx.BadRequest(c, "invalid json")
			return
		}
		log.Printf("Create web app request: %s\n", webAppDTO.String())

		if err := dto.Validate.Struct(&webAppDTO); err != nil {
			handleValidationError(err, c)
			return
		}

		wa, err := mapper.DtoToWebapp(webAppDTO)
		if err != nil {
			log.Printf("failed to convert dto to webapp")
			httpx.BadRequest(c, err.Error())
			return
		}

		id, err := s.Insert(c.Request.Context(), *wa)
		if err != nil {
			log.Printf("failed to insert web app to db: %v", err)
			httpx.InternalError(c)
		}
		wa.ID = id

		err = manager.SetProxyToManager(wa)
		if err != nil {
			log.Printf("failed to add proxy to manager: %s", err)
			httpx.InternalError(c)
			return
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

		var waDTO webappDto.WebappDto
		if err := c.ShouldBindJSON(&waDTO); err != nil {
			httpx.BadRequest(c, "invalid json body")
			return
		}

		if err := dto.Validate.Struct(&waDTO); err != nil {
			handleValidationError(err, c)
			return
		}

		wa, err := mapper.DtoToWebapp(waDTO)
		if err != nil {
			httpx.BadRequest(c, err.Error())
			return
		}
		wa.ID = id

		if err := s.Edit(c.Request.Context(), wa); err != nil {
			log.Printf("failed to edit web app: %v", err)
			httpx.InternalError(c)
			return
		}

		err = manager.SetProxyToManager(wa)
		if err != nil {
			log.Printf("failed to add proxy to manager: %s", err)
			httpx.InternalError(c)
			return
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
		wa, err := s.FindById(c.Request.Context(), id)
		if err != nil {
			if errors.Is(err, mongorepository.ErrNotFound) {
				httpx.NotFound(c, "web app "+wa.Name)
				return
			}
			httpx.InternalError(c)
			return
		}
		if err = s.Delete(c.Request.Context(), wa); err != nil {
			log.Printf("failed to delete web app %v: %v", wa.Name, err)
			httpx.InternalError(c)
			return
		}

		manager.DeleteProxyFromManager(wa)

		c.Status(204)
	}
}

func handleValidationError(err error, c *gin.Context) {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		log.Println("Validation error!!!")
		for _, fe := range ve {
			log.Printf("validation field[%s] failed: [%s], value[%v]", fe.Field(), fe.Tag(), fe.Value())
		}
		httpx.ValidationError(c, ve)
		return
	}
	log.Println("invalid request")
	httpx.BadRequest(c, "invalid request")
	return
}
