package dto

import (
	policy "reverseproxy/internal/domain/policy"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PolicyDTO struct {
	ID   string   `json:"id" validate:"omitempty,hexadecimal,len=24"`
	Name string   `json:"name" validate:"required,min=3,max=64"`
	WL   []string `json:"wl" validate:"dive,wl"` // TO DO валидатор + подумать, что будет если он будет пустой
}

func PolicyToDTO(p *policy.Policy) *PolicyDTO {
	return &PolicyDTO{
		ID:   p.ID.Hex(),
		Name: p.Name,
		WL:   p.WL,
	}
}

func DTOToPolicy(dto PolicyDTO) (*policy.Policy, error) {
	if dto.ID != "" {
		id, err := primitive.ObjectIDFromHex(dto.ID)
		if err != nil {
			return nil, err
		}
		return &policy.Policy{
			ID:   id,
			Name: dto.Name,
			WL:   dto.WL,
		}, nil
	} else {
		return &policy.Policy{
			Name: dto.Name,
			WL:   dto.WL,
		}, nil
	}
}
