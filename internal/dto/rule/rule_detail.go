package rule

import (
	"reverseproxy/internal/domain/action"
	"reverseproxy/internal/domain/policy"
	"reverseproxy/internal/domain/rule"
)

type RuleDetailResponse struct {
	Rule              RuleDetailView    `json:"rule"`
	AvailableActions  []ActionParamView `json:"available_actions"`
	AvailablePolicies []ShortPolicyView `json:"available_policies"`
}

type RuleDetailView struct {
	ID                 string                  `json:"id"`
	Name               string                  `json:"name"`
	Enabled            bool                    `json:"enabled"`            // Включено ли правило
	Actions            []ActionParamView       `json:"actions"`            // Список действий для правила
	Policies           []string                `json:"policies"`           // Список использующихся политик
	PolicyActionParams []PolicyActionParamView `json:"policyActionParams"` // Список переопределений действий для политик
	ExprView           ExprView                `json:"expr"`               // Само выражение правила
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

type ShortPolicyView struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func BuildRuleResponse(ruleDetail *rule.RuleDetail, actions []action.ActionDoc, policies []policy.Policy) (*RuleDetailResponse, error) {
	availableActions := make([]ActionParamView, 0, len(actions))
	for _, a := range actions {
		availableActions = append(availableActions, ActionParamView{
			ID:   a.ID.Hex(),
			Name: a.Name,
		})
	}

	availablePolicies := make([]ShortPolicyView, 0, len(policies))
	for _, p := range policies {
		availablePolicies = append(availablePolicies, ShortPolicyView{
			ID:   p.ID.Hex(),
			Name: p.Name,
		})
	}

	return &RuleDetailResponse{
		Rule:              *ToRuleDetailView(ruleDetail, policies),
		AvailableActions:  availableActions,
		AvailablePolicies: availablePolicies,
	}, nil
}

func ToRuleDetailView(detail *rule.RuleDetail, policies []policy.Policy) *RuleDetailView {
	rdv := RuleDetailView{
		ID:                 detail.ID.Hex(),
		Name:               detail.Name,
		Enabled:            detail.Enabled,
		Actions:            make([]ActionParamView, 0, len(detail.Actions)),
		PolicyActionParams: make([]PolicyActionParamView, 0, len(detail.PolicyActionParams)),
	}

	rdv.Policies = make([]string, 0, len(detail.Policies))
	for _, policyId := range detail.Policies {
		for _, policy := range policies {
			if policy.ID == policyId {
				rdv.Policies = append(rdv.Policies, policy.Name)
				break
			}
		}
	}

	for _, actionParam := range detail.Actions {
		rdv.Actions = append(rdv.Actions, ActionParamView{
			ID:   actionParam.ID.Hex(),
			Name: actionParam.Name,
		})
	}

	for _, policyActionParam := range detail.PolicyActionParams {
		papv := PolicyActionParamView{
			ID:      policyActionParam.ID.Hex(),
			Name:    policyActionParam.Name,
			Actions: make([]ActionParamView, 0),
		}
		for _, actionParam := range policyActionParam.Actions {
			papv.Actions = append(papv.Actions, ActionParamView{
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

func buildExprView(doc rule.ExprDoc) ExprView {
	exprView := ExprView{
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
