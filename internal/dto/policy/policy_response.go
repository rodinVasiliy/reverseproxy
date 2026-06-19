package policy

type ListItem struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	WL      []string `json:"wl"`
	Webapps []string `json:"webapps"`
}

type Detail struct {
	ID    string        `json:"id"`
	Name  string        `json:"name"`
	WL    []string      `json:"wl"`
	Rules []RuleRefView `json:"rules"`
}

type RuleRefView struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}
