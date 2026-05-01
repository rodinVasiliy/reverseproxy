package rule

type RuleDetailView struct {
	ID                 string                  `json:"id"`      // Rule ID
	Name               string                  `json:"name"`    // Rule Name
	Enabled            bool                    `json:"enabled"` // Is Rule enabled
	Actions            []ActionParamView       `json:"actions"`
	PolicyActionParams []PolicyActionParamView `json:"policyActionParams"`
	ExprView           ExprView                `json:"expr"`
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

type ExprView struct {
	NodeType string     `json:"nodeType"` // "condition" | "group"
	IsNot    bool       `json:"isNot"`    // Если выставлено - значит при true вернет false и наоборот
	Operator string     `json:"operator"` //
	Children []ExprView `json:"children"` // Если группа
	Match    string     `json:"match"`    // equals/in/regex
	Field    string     `json:"field"`    // Поле которое будет проверяться, например "ua"
	Raw      string     `json:"value"`    // Значение на которое будет матчится параметр запроса
}
