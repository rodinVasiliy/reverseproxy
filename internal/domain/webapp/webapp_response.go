package webapp

type Response struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	PolicyId   string `json:"policyId"`
	PolicyName string `json:"policyName"`

	SSLId   string `json:"sslId"`
	SSLName string `json:"sslName"`

	Port     int      `json:"port"`
	Upstream string   `json:"upstream"`
	Hosts    []string `json:"hosts"`
}
