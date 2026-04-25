package mapper

import (
	"reverseproxy/internal/domain/action"
	responseAction "reverseproxy/internal/dto/action"
)

func ToActionResponse(doc action.ActionDoc) *responseAction.ActionResponse {
	return &responseAction.ActionResponse{
		ID:   doc.ID.Hex(),
		Name: doc.Name,
	}
}
