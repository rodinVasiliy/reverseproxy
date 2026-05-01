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

func ToRuleResponse(detail *rule.RuleDetail) *ruleDto.RuleDetailView {
	rdv := ruleDto.RuleDetailView{
		ID:                 detail.ID.Hex(),
		Name:               detail.Name,
		Enabled:            detail.Enabled,
		Actions:            make([]ruleDto.ActionParamView, 0, len(detail.Actions)),
		PolicyActionParams: make([]ruleDto.PolicyActionParamView, 0, len(detail.PolicyActionParams)),
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
		for _, actionParam := range detail.Actions {
			papv.Actions = append(papv.Actions, ruleDto.ActionParamView{
				ID:   actionParam.ID.Hex(),
				Name: actionParam.Name,
			})
		}
		rdv.PolicyActionParams = append(rdv.PolicyActionParams, papv)
	}

	exprView := buildExprView(*detail.Expr)
	rdv.ExprView = exprView

	return &rdv
}

func buildExprView(doc rule.ExprDoc) ruleDto.ExprView {
	exprView := ruleDto.ExprView{
		NodeType: doc.NodeType,
		IsNot:    doc.IsNot,
		Operator: doc.Operator,
		Match:    doc.Match,
		Field:    doc.Field,
		Raw:      doc.Raw,
	}

	for _, child := range doc.Children {
		exprView.Children = append(exprView.Children, buildExprView(child))
	}

	return exprView
}
