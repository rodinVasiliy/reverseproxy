package rule

type RuleMetaResponse struct {
	AvailableActions  []ActionParamView `json:"available_actions"`
	AvailablePolicies []ShortPolicyView `json:"available_policies"`
}
