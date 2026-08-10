package rule

// TODO валидацию написать
type RuleDto struct {
	Name            string            `json:"name" validate:"required,min=3,max=64"`
	Enabled         bool              `json:"enabled"`                  // Включено ли правило
	Actions         []string          `json:"actions"`                  // Список действий, которые будут применяться при срабатывании правила
	Policies        []string          `json:"policies"`                 // Список политик, для которых это правило будет работать
	PolicyOverrides []PolicyOverrides `json:"overrides"`                // Список переопределений, для политики переопределяется набор действий
	Expression      ExprDto           `json:"expr" validate:"required"` // Само правило
}

type PolicyOverrides struct { // Переопределяет для правила с ID список действий
	ID      string   `json:"id"`      // policyId
	Actions []string `json:"actions"` // actionIds
}

type ExprDto struct {
	NodeType string    `json:"nodeType"`           // "condition" | "group"
	IsNot    bool      `json:"isNot"`              // Если выставлено - значит при true вернет false и наоборот
	Operator string    `json:"operator,omitempty"` // AND или OR который будет использоваться между Conditions у группы
	Children []ExprDto `json:"children,omitempty"` // Если группа
	Match    string    `json:"match,omitempty"`    // equals/in/regex
	Field    string    `json:"field,omitempty"`    // Поле которое будет проверяться, например "ua"
	Value    string    `json:"value,omitempty"`    // Значение на которое будет матчится параметр запроса
}
