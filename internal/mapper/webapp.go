package mapper

import (
	"reverseproxy/internal/domain/webapp"
	webappDto "reverseproxy/internal/dto/webapp"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func WebappToDto(webapp webapp.WebApp) *webappDto.WebappDto {
	return &webappDto.WebappDto{
		ID:       webapp.ID.Hex(),
		Name:     webapp.Name,
		PolicyId: webapp.PolicyId.Hex(),
		Port:     webapp.Port,
		SSLId:    webapp.SSLId.Hex(),
		Upstream: webapp.Upstream,
		Hosts:    webapp.Hosts,
	}
}

func DtoToWebapp(dto webappDto.WebappDto) (*webapp.WebApp, error) {
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
