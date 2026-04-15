package policy

type Response struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	WL      []string `json:"wl"`
	Webapps []string `json:"webapps"`
}
