package ssl

type SSLConfigurationDTO struct {
	ID           string `json:"id" validate:"omitempty,hexadecimal,len=24"`
	Name         string `json:"name" validate:"required,min=3,max=64"`
	CertFileName string `json:"cert" validate:"required,certfilename"`
	KeyFileName  string `json:"key" validate:"required,keyfilename"`
}
