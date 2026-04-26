package rule

type RuleDetail struct {
	ID                 string                  `json:"id"`
	Name               string                  `json:"name"`
	Enabled            bool                    `json:"enabled"`
	Actions            []ActionParamView       `json:"actions"`
	PolicyActionParams []PolicyActionParamView `json:"policyActionParams"`
}

type ActionParamView struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type PolicyActionParamView struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Actions []ActionParamView `json:"actions"`
}
