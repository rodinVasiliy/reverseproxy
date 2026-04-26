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

func ToPolicyResponse(detail *rule.RuleDetail) *ruleDto.RuleDetailView {
	rdv := ruleDto.RuleDetailView{
		ID:                 detail.ID.Hex(),
		Name:               detail.Name,
		Enabled:            detail.Enabled,
		Actions:            make([]ruleDto.ActionParamView, 0),
		PolicyActionParams: make([]ruleDto.PolicyActionParamView, 0),
	}

	for _, actionParam := range detail.Actions {
		rdv.Actions = append(rdv.Actions, ruleDto.ActionParamView{
			ID:   actionParam.ID.Hex(),
			Name: actionParam.Name,
		})
	}

	for _, policyActionParam := range detail.PolicyActionParams {
		papv := ruleDto.PolicyActionParamView{
			ID:      policyActionParam.ID.Hex(),
			Name:    policyActionParam.Name,
			Actions: make([]ruleDto.ActionParamView, 0),
		}
		for _, actionParam := range detail.PolicyActionParams {
			papv.Actions = append(papv.Actions, ruleDto.ActionParamView{
				ID:   actionParam.ID.Hex(),
				Name: actionParam.Name,
			})
		}
		rdv.PolicyActionParams = append(rdv.PolicyActionParams, papv)
	}

	return &rdv
}
