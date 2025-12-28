package dto

import (
	webapp "reverseproxy/internal/model/webapp"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TODO - скорректировать валидацию
type WebAppDTO struct {
	ID       string   `json:"id" validate:"omitempty,hexadecimal,len=24"`
	Name     string   `json:"name" validate:"required"`
	PolicyId string   `json:"policyId" validate:"omitempty,hexadecimal,len=24"`
	Port     int      `json:"port" validate:"min=1,max=65535"`
	SSLId    string   `json:"sslId" validate:"omitempty,hexadecimal,len=24"`
	Upstream string   `json:"upstream" validate:"required"`
	Hosts    []string `json:"hosts" validate:"min=1,dive,hostname"`
}

func WebAppToDTO(webapp webapp.WebApp) *WebAppDTO {
	return &WebAppDTO{
		ID:       webapp.ID.Hex(),
		Name:     webapp.Name,
		PolicyId: webapp.PolicyId.Hex(),
		Port:     webapp.Port,
		SSLId:    webapp.SSLId.Hex(),
		Upstream: webapp.Upstream,
		Hosts:    webapp.Hosts,
	}
}

func DTOToWebApp(dto WebAppDTO) (*webapp.WebApp, error) {
	policyID, err := primitive.ObjectIDFromHex(dto.PolicyId)
	if err != nil {
		return nil, err
	}
	sslID, err := primitive.ObjectIDFromHex(dto.SSLId)
	if err != nil {
		return nil, err
	}
	return &webapp.WebApp{
		Name:     dto.Name,
		PolicyId: policyID,
		SSLId:    sslID,
		Hosts:    dto.Hosts,
		Upstream: dto.Upstream,
		Port:     dto.Port,
	}, nil
}
