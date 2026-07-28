package policy

type Dto struct {
	ID   string   `json:"id" validate:"omitempty,hexadecimal,len=24"`
	Name string   `json:"name" validate:"required,min=3,max=64"`
	WL   []string `json:"wl" validate:"dive,required,wl"`
}
