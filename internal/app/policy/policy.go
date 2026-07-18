package policy

import (
	"context"
	"reverseproxy/internal/domain/action"
	"reverseproxy/internal/domain/policy"
	"reverseproxy/internal/domain/rule"
	"reverseproxy/internal/domain/webapp"
	policyDto "reverseproxy/internal/dto/policy"
	engine "reverseproxy/internal/waf/policy"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AppPolicyService struct {
	policyService *policy.Service
	actionService *action.Service
	ruleService   *rule.Service
	webappService *webapp.Service

	policyEngine *engine.PolicyEngine
}

func NewAppPolicyService(p *policy.Service, a *action.Service, r *rule.Service, w *webapp.Service, pe *engine.PolicyEngine) *AppPolicyService {
	return &AppPolicyService{
		policyService: p,
		actionService: a,
		ruleService:   r,
		webappService: w,
		policyEngine:  pe,
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

func (s *AppPolicyService) GetPolicyDetailById(ctx context.Context, id primitive.ObjectID) (*policy.Detail, error) {
	p, err := s.policyService.FindById(ctx, id)
	if err != nil {
		return nil, err
	}

	rules, err := s.ruleService.FindByPolicyId(ctx, p.ID)
	if err != nil {
		return nil, err
	}

	detail := policy.Detail{
		ID:    id,
		Name:  p.Name,
		WL:    p.WL,
		Rules: make([]policy.RuleDetail, 0, len(rules)),
	}
	for _, rule := range rules {
		detail.Rules = append(detail.Rules, policy.RuleDetail{
			ID:      rule.ID,
			Name:    rule.Name,
			Enabled: rule.Enabled,
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

func (s *AppPolicyService) Update(ctx context.Context, ps *policy.Service, policy *policy.Policy) error {
	err := ps.Update(ctx, policy)
	if err != nil {
		return err
	}

	err = s.policyEngine.Update(*policy)
	if err != nil {
		return err
	}

	return nil
}

func (s *AppPolicyService) Delete(ctx context.Context, p *policy.Policy, ps *policy.Service) error {
	err := ps.Delete(ctx, p)
	if err != nil {
		return err
	}

	s.policyEngine.Delete(*p)
	return nil
}

func (s *AppPolicyService) Create(ctx context.Context, dto policyDto.Dto) error {
	p := policy.Policy{
		Name: dto.Name,
		WL:   dto.WL,
	}
	return s.policyService.Create(ctx, &p)
}

func sliceToMap[T any, K comparable](items []T, keyFn func(T) K) map[K]T {
	result := make(map[K]T, len(items))
	for _, item := range items {
		result[keyFn(item)] = item
	}
	return result
}
