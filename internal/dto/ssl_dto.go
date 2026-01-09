package dto

import (
	ssl "reverseproxy/internal/domain/ssl"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SSLConfigurationDTO struct {
	ID           string `json:"id" validate:"omitempty,hexadecimal,len=24"`
	Name         string `json:"name" validate:"required,min=3,max=64"`
	CertFileName string `json:"cert" validate:"required,certfilename"`
	KeyFileName  string `json:"key" validate:"required,keyfilename"`
}

func ToSSLConfigDTO(ssl ssl.SSLConfiguration) *SSLConfigurationDTO {
	return &SSLConfigurationDTO{
		ID:           ssl.ID.Hex(),
		Name:         ssl.Name,
		CertFileName: ssl.CertFileName,
		KeyFileName:  ssl.KeyFileName,
	}
}

// Вернет ошибку, если не смог сконвертировать id в primitive.ObjectId
func DTOToSSLConfig(dto SSLConfigurationDTO) (*ssl.SSLConfiguration, error) {
	SSLConfiguration := ssl.SSLConfiguration{
		Name:         dto.Name,
		CertFileName: dto.CertFileName,
		KeyFileName:  dto.KeyFileName,
	}

	if dto.ID != "" {
		id, err := primitive.ObjectIDFromHex(dto.ID)
		if err != nil {
			return nil, err
		}
		SSLConfiguration.ID = id
	}
	return &SSLConfiguration, nil
}
