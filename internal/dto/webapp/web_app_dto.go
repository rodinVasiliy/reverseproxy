package webapp

import (
	"fmt"
	"strings"
)

type WebappDto struct {
	ID       string   `json:"id" validate:"omitempty,hexadecimal,len=24"`
	Name     string   `json:"name" validate:"required,webappname"`
	PolicyId string   `json:"policyId" validate:"omitempty,hexadecimal,len=24"`
	Port     int      `json:"port" validate:"min=1,max=65535"`
	SSLId    string   `json:"sslId" validate:"omitempty,hexadecimal,len=24"`
	Upstream string   `json:"upstream" validate:"required,upstream"`
	Hosts    []string `json:"hosts" validate:"min=1,dive,host"` // Dive = проверять каждый элемент
}

func (dto WebappDto) String() string {
	var builder strings.Builder
	for _, host := range dto.Hosts {
		builder.WriteString(host + " ")
	}
	return fmt.Sprintf("name:%s; port:%v; upstream:%s; policyId:%s; sslId:%s, hosts %s",
		dto.Name, dto.Port, dto.Upstream, dto.PolicyId, dto.SSLId, builder.String())
}
