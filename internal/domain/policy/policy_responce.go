package policy

type Response struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Webapps []string `json:"webapps"`
}
