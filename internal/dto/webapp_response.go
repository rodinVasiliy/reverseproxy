package dto

type WebAppResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	Port     int    `json:"port"`
	Upstream string `json:"upstream"`

	PolicyId   string `json:"policyId"`
	PolicyName string `json:"policyName"`

	SSLId   string `json:"sslId"`
	SSLName string `json:"sslName"`
}
