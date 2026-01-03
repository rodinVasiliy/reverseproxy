package handler

import (
	"errors"
	"log"
	dto "reverseproxy/internal/dto"
	httpx "reverseproxy/internal/httpx"
	ssl "reverseproxy/internal/model/ssl"
	mongorep "reverseproxy/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func RegisterSSLRoutes(r *gin.RouterGroup, s *ssl.Service) {
	g := r.Group("/ssls")
	{
		g.GET("", getSSLConfigs(s))
		g.POST("", createSSLConfig(s))
		g.DELETE("/:id", deleteSSLConfig(s))
		g.PUT("/:id", updateSSLConfig(s))
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

		dtos := make([]dto.SSLConfigurationDTO, 0, len(sslConfigs))
		for _, sslConfig := range sslConfigs {
			dto := dto.ToSSLConfigDTO(sslConfig)
			dtos = append(dtos, *dto)
		}
		c.JSON(200, dtos)
	}
}

func deleteSSLConfig(s *ssl.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			httpx.BadRequest(c, "invalid id")
			return
		}

		ssl, err := s.FindByID(c.Request.Context(), id)
		if err != nil {
			if errors.Is(err, mongorep.ErrNotFound) {
				httpx.NotFound(c, "ssl_config")
				return
			}
			log.Printf("failed to find ssl by id: %v, %v", id.Hex(), err)
			httpx.InternalError(c)
			return
		}

		err = s.Delete(c.Request.Context(), ssl)
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
			httpx.BadRequest(c, "invalid json body")
			return
		}

		if err := dto.Validate.Struct(sslDTO); err != nil {
			var ve validator.ValidationErrors
			if errors.As(err, &ve) {
				httpx.ValidationError(c, ve)
			}
			httpx.BadRequest(c, "invalid request")
			return
		}

		ssl, err := dto.DTOToSSLConfig(sslDTO)
		if err != nil {
			httpx.BadRequest(c, err.Error())
			return
		}

		_, err = s.Insert(c.Request.Context(), *ssl)
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

		ssl, err := dto.DTOToSSLConfig(sslDTO)
		if err != nil {
			httpx.BadRequest(c, err.Error())
			return
		}
		ssl.ID = id
		err = s.Update(c.Request.Context(), ssl)
		if err != nil {
			log.Printf("failed to update ssl config: %v", err)
			// TODO мб добавить сюда отлов ErrNotFound + в Update засунуть в описание метода то, что может вернуть эту ошибку
			httpx.InternalError(c)
			return
		}

		c.Status(204)
	}
}
