package rule

// TODO валидацию написать
type RuleDto struct {
	Name      string      `json:"name"`
	Enabled   bool        `json:"enabled"`
	Actions   []string    `json:"actions"`
	Overrides []Overrides `json:"overrides"`
}

type Overrides struct {
	ID      string   `json:"id"`      // policyId
	Actions []string `json:"actions"` // actionIds
}
