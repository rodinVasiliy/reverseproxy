package handler

import (
	"errors"
	"log"
	"net/http"
	"os"
	"reverseproxy/internal/domain/ssl"
	"reverseproxy/internal/domain/webapp"
	"reverseproxy/internal/dto"
	"reverseproxy/internal/httpx"
	mongorep "reverseproxy/internal/infrastructure/mongo"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func RegisterSSLRoutes(r *gin.RouterGroup, s *ssl.Service, w *webapp.Service) {
	g := r.Group("/ssls")
	{
		g.GET("", getSSLConfigs(s))
		g.POST("", createSSLConfig(s))
		g.DELETE("/:id", deleteSSLConfig(s, w))
		g.PUT("/:id", updateSSLConfig(s))
		g.GET("/files", listSSLFiles())
		g.GET("/:id", getSSLConfigByID(s))
	}
}

func getSSLConfigs(s *ssl.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		sslConfigs, err := s.FindAll(c.Request.Context())
		if err != nil {
			log.Printf("failed to find ssl configs: %v", err)
			httpx.InternalError(c)
			return
		}

		sslConfigurationDTOS := make([]dto.SSLConfigurationDTO, 0, len(sslConfigs))
		for _, sslConfig := range sslConfigs {
			sslConfigDTO := dto.ToSSLConfigDTO(sslConfig)
			sslConfigurationDTOS = append(sslConfigurationDTOS, *sslConfigDTO)
		}
		c.JSON(200, sslConfigurationDTOS)
	}
}

func getSSLConfigByID(s *ssl.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			log.Printf("failed to parse id: %v", err)
			httpx.BadRequest(c, "invalid id")
			return
		}
		sslConfig, err := s.FindByID(c.Request.Context(), id)
		if err != nil {
			log.Printf("failed to find ssl config by id: %v", err)
			httpx.InternalError(c)
		}
		sslConfigDTO := dto.ToSSLConfigDTO(*sslConfig)
		c.JSON(200, sslConfigDTO)
	}
}

func deleteSSLConfig(s *ssl.Service, w *webapp.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			httpx.BadRequest(c, "invalid id")
			return
		}

		sslConfiguration, err := s.FindByID(c.Request.Context(), id)
		if err != nil {
			if errors.Is(err, mongorep.ErrNotFound) {
				httpx.NotFound(c, "ssl_config")
				return
			}
			log.Printf("failed to find ssl by id: %v, %v", id.Hex(), err)
			httpx.InternalError(c)
			return
		}

		webapps, err := w.FindBySSLId(id, c.Request.Context())
		if err != nil {
			log.Printf("failed to find webapps by id: %v, %v", id.Hex(), err)
			httpx.InternalError(c)
		}
		if len(webapps) > 0 {
			names := make([]string, 0, len(webapps))
			for _, w := range webapps {
				names = append(names, w.Name)
			}
			log.Printf("ssl %v in use in webapps: %v", sslConfiguration.Name, names)

			c.JSON(http.StatusBadRequest, gin.H{
				"code":    "ssl_in_use",
				"message": "SSL config is used",
				"webapps": names,
			})
			return
		}
		err = s.Delete(c.Request.Context(), sslConfiguration)
		if err != nil {
			log.Printf("failed to delete ssl config: %v", err)
			httpx.InternalError(c)
			return
		}

		c.Status(204)
	}
}

func createSSLConfig(s *ssl.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var sslDTO dto.SSLConfigurationDTO
		if err := c.ShouldBindJSON(&sslDTO); err != nil {
			log.Println("invalid json body: ", err)
			httpx.BadRequest(c, "invalid json body")
			return
		}

		if err := dto.Validate.Struct(sslDTO); err != nil {
			var ve validator.ValidationErrors
			if errors.As(err, &ve) {
				log.Println("validation error: ", err)
				httpx.ValidationError(c, ve)
			}
			log.Println("invalid request: ", err)
			httpx.BadRequest(c, "invalid request")
			return
		}

		sslConfig, err := dto.DTOToSSLConfig(sslDTO)
		if err != nil {
			log.Println("failed to convert DTO to SSL Config: ", err)
			httpx.BadRequest(c, err.Error())
			return
		}

		_, err = s.Insert(c.Request.Context(), *sslConfig)
		if err != nil {
			log.Printf("failed to insert ssl config: %v", err)
			httpx.InternalError(c)
		}

		c.Status(201)
	}
}

func updateSSLConfig(s *ssl.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			httpx.BadRequest(c, "invalid id")
			return
		}

		var sslDTO dto.SSLConfigurationDTO
		if err = c.ShouldBindJSON(&sslDTO); err != nil {
			httpx.BadRequest(c, "invalid json")
			return
		}

		if err := dto.Validate.Struct(sslDTO); err != nil {
			var ve validator.ValidationErrors
			if errors.As(err, &ve) {
				httpx.ValidationError(c, ve)
				return
			}
			httpx.BadRequest(c, "invalid request")
			return
		}

		sslConfig, err := dto.DTOToSSLConfig(sslDTO)
		if err != nil {
			httpx.BadRequest(c, err.Error())
			return
		}
		sslConfig.ID = id
		err = s.Update(c.Request.Context(), sslConfig)
		if err != nil {
			log.Printf("failed to update ssl config: %v", err)
			// TODO мб добавить сюда отлов ErrNotFound + в Update засунуть в описание метода то, что может вернуть эту ошибку
			httpx.InternalError(c)
			return
		}

		c.Status(204)
	}
}

func listSSLFiles() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// файлы с сертификатами и ключами у нас находятся тут
		files, err := os.ReadDir("/etc/nginx/ssl")
		if err != nil {
			log.Printf("failed to read ssl directory")
			httpx.InternalError(ctx)
			return
		}

		var certs []string
		var keys []string

		for _, f := range files {
			if f.IsDir() {
				continue
			}

			name := f.Name()

			// TO DO сравнить с валидацией
			if strings.HasSuffix(name, ".pem") || strings.HasSuffix(name, ".cer") || strings.HasSuffix(name, ".crt") {
				certs = append(certs, name)
			}

			if strings.HasSuffix(name, ".key") || strings.HasSuffix(name, ".pem") {
				keys = append(keys, name)
			}
		}

		ctx.JSON(200, gin.H{
			"certs": certs,
			"keys":  keys,
		})
	}
}
