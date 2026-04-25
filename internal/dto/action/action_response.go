package action

type ActionResponse struct {
	ID   string `json:"id" validate:"omitempty,hexadecimal,len=24"`
	Name string `json:"name" validate:"required,min=3,max=64"`
}
