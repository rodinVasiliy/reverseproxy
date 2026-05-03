package mapper

import (
	"reverseproxy/internal/domain/rule"
	ruleDto "reverseproxy/internal/dto/rule"
)

func ToRuleListItems(rules []rule.Rule) []ruleDto.RuleListItem {
	items := make([]ruleDto.RuleListItem, len(rules))
	for i, item := range rules {
		items[i] = ruleDto.RuleListItem{
			ID:      item.ID.Hex(),
			Name:    item.Name,
			Enabled: item.Enabled,
		}
	}
	return items
}
