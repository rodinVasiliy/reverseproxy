package policy

import (
	"reverseproxy/internal/waf/action"
	"reverseproxy/internal/waf/rule"
)

type PolicyRule struct {
	Rule    *rule.CompiledRule
	Actions []action.Action
}
