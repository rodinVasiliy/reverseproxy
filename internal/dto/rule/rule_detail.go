package rule

type RuleDetailView struct {
	ID                 string                  `json:"id"`      // Rule ID
	Name               string                  `json:"name"`    // Rule Name
	Enabled            bool                    `json:"enabled"` // Is Rule enabled
	Actions            []ActionParamView       `json:"actions"`
	PolicyActionParams []PolicyActionParamView `json:"policyActionParams"`
}

type ActionParamView struct {
	ID   string `json:"id"`   // Action ID
	Name string `json:"name"` // Action name
}

type PolicyActionParamView struct {
	ID      string            `json:"id"`   // Policy ID
	Name    string            `json:"name"` // Policy Name
	Actions []ActionParamView `json:"actions"`
}
