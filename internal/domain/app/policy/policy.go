package policy

import (
	"context"
	"log"
	"reverseproxy/internal/domain/action"
	"reverseproxy/internal/domain/policy"
	"reverseproxy/internal/domain/rule"
	"reverseproxy/internal/domain/webapp"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AppPolicyService struct {
	policyService *policy.Service
	actionService *action.Service
	ruleService   *rule.Service
	webappService *webapp.Service
}

func NewAppPolicyService(p *policy.Service, a *action.Service, r *rule.Service, w *webapp.Service) *AppPolicyService {
	return &AppPolicyService{
		policyService: p,
		actionService: a,
		ruleService:   r,
		webappService: w,
	}
}

func (s *AppPolicyService) GetWebappsByPolicyId(ctx context.Context, id primitive.ObjectID) ([]string, error) {
	return s.webappService.FindByPolicyId(id, ctx)
}

func (s *AppPolicyService) List(ctx context.Context) ([]policy.PolicyListItem, error) {
	policies, err := s.policyService.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]policy.PolicyListItem, 0, len(policies))

	policyIds := make([]primitive.ObjectID, len(policies))
	for i := range policies {
		policyIds[i] = policies[i].ID
	}

	webappsMap, err := s.webappService.FindByPolicyIDs(policyIds, ctx)
	if err != nil {
		return nil, err
	}

	for _, p := range policies {
		items = append(items, policy.PolicyListItem{
			ID:      p.ID,
			Name:    p.Name,
			WL:      p.WL,
			Webapps: webappsMap[p.ID],
		})
	}
	return items, nil
}

func (s *AppPolicyService) GetPolicyDetailById(ctx context.Context, id primitive.ObjectID) (*policy.PolicyDetail, error) {
	p, err := s.policyService.FindById(ctx, id)
	if err != nil {
		return nil, err
	}

	actions, err := s.actionService.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	actionsMap := sliceToMap(actions, func(a action.ActionDoc) primitive.ObjectID {
		return a.ID
	})

	rules, err := s.ruleService.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	rulesMap := sliceToMap(rules, func(r rule.Rule) primitive.ObjectID {
		return r.ID
	})

	var detail policy.PolicyDetail
	detail.ID = id
	detail.Name = p.Name
	detail.WL = p.WL
	detail.Rules = make([]policy.PolicyRuleDetail, 0, len(rules))
	log.Printf("found %d rules\n", len(rules))

	for _, rr := range p.Rules {
		r := rulesMap[rr.RuleID]

		var actionIDs []primitive.ObjectID
		if len(rr.Actions) > 0 {
			actionIDs = rr.Actions
		} else {
			actionIDs = r.Actions
		}

		var actionViews []policy.ActionDetail
		for _, aid := range actionIDs {
			a := actionsMap[aid]
			actionViews = append(actionViews, policy.ActionDetail{
				ID:   a.ID,
				Name: a.Name,
			})
		}

		detail.Rules = append(detail.Rules, policy.PolicyRuleDetail{
			ID:      r.ID,
			Name:    r.Name,
			Enabled: r.Enabled,
			Actions: actionViews,
		})

	}

	return &detail, nil

}

func (s *AppPolicyService) CanDeletePolicy(ctx context.Context, id primitive.ObjectID) error {
	webapps, err := s.webappService.FindByPolicyId(id, ctx)
	if err != nil {
		return err
	}
	if len(webapps) > 0 {
		return &policy.PolicyInUseError{Webapps: webapps}
	}
	return nil
}

func sliceToMap[T any, K comparable](items []T, keyFn func(T) K) map[K]T {
	result := make(map[K]T, len(items))
	for _, item := range items {
		result[keyFn(item)] = item
	}
	return result
}
