package mapper

import (
	"reverseproxy/internal/domain/ssl"
	ssl2 "reverseproxy/internal/dto/ssl"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func ToSSLConfigDTO(ssl ssl.SSL) *ssl2.SSLConfigurationDTO {
	return &ssl2.SSLConfigurationDTO{
		ID:           ssl.ID.Hex(),
		Name:         ssl.Name,
		CertFileName: ssl.CertFileName,
		KeyFileName:  ssl.KeyFileName,
	}
}

func DTOToSSLConfig(dto ssl2.SSLConfigurationDTO) (*ssl.SSL, error) {
	SSLConfiguration := ssl.SSL{
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
