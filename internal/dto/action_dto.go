package dto

import (
	action "reverseproxy/internal/model/action"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ActionDTO struct {
	ID   string `json:"id" validate:"omitempty,hexadecimal,len=24"`
	Name string `json:"name" validate:"required,min=3,max=64"`
}

func DTOToAction(dto *ActionDTO) (*action.ActionDoc, error) {
	// если id пустое, ничего не надо приводить к primitive.ObjectID, если не пустое - приводим и проверяем на ошибку
	if dto.ID != "" {
		id, err := primitive.ObjectIDFromHex(dto.ID)
		if err != nil {
			return nil, err
		}
		return &action.ActionDoc{
			ID:   id,
			Name: dto.Name,
		}, nil
	} else {
		return &action.ActionDoc{
			Name: dto.Name,
		}, nil
	}
}

func ActionToDTO(action *action.ActionDoc) *ActionDTO {
	return &ActionDTO{
		ID:   action.ID.Hex(),
		Name: action.Name,
	}
}
