package policy

import (
	"context"
	"reverseproxy/internal/domain/action"
	"reverseproxy/internal/domain/rule"
	repository "reverseproxy/internal/infrastructure/mongo"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Service struct {
	repository     *repository.MongoRepository[Policy]
	actionService  *action.Service
	ruleService    *rule.Service
	webappProvider WebappProvider
}

func NewService(repo *repository.MongoRepository[Policy]) *Service {
	return &Service{repository: repo}
}

func (s *Service) SetRuleService(srv *rule.Service) {
	s.ruleService = srv
}

func (s *Service) SetActionService(actionService *action.Service) {
	s.actionService = actionService
}

func (s *Service) SetWebappProvider(webappProvider WebappProvider) {
	s.webappProvider = webappProvider
}

func (s *Service) Insert(ctx context.Context, policy Policy) (primitive.ObjectID, error) {
	return s.repository.Insert(ctx, policy)
}

func (s *Service) FindByName(ctx context.Context, name string) (*Policy, error) {
	return s.repository.FindOne(ctx, bson.M{"name": name})
}

func (s *Service) FindById(ctx context.Context, id primitive.ObjectID) (*Policy, error) {
	return s.repository.FindById(ctx, id)
}

func (s *Service) FindAll(ctx context.Context) ([]Policy, error) {
	return s.repository.FindAll(ctx)
}

func (s *Service) GetPolicyDetailById(ctx context.Context, id primitive.ObjectID) (*PolicyDetail, error) {
	policy, err := s.FindById(ctx, id)
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

	var detail PolicyDetail
	detail.ID = id
	detail.Name = policy.Name
	detail.WL = policy.WL

	for _, rr := range policy.Rules {
		rule := rulesMap[rr.RuleID]

		var actionIDs []primitive.ObjectID
		if len(rr.Actions) > 0 {
			actionIDs = rr.Actions
		} else {
			actionIDs = rule.Actions
		}

		var actionViews []ActionDetail
		for _, aid := range actionIDs {
			a := actionsMap[aid]
			actionViews = append(actionViews, ActionDetail{
				ID:   a.ID,
				Name: a.Name,
			})
		}

		detail.Rules = append(detail.Rules, PolicyRuleDetail{
			ID:      rule.ID,
			Name:    rule.Name,
			Enabled: rule.Enabled,
			Actions: actionViews,
		})

	}

	return &detail, nil

}

func (s *Service) GetWebappsByPolicyId(ctx context.Context, id primitive.ObjectID) ([]string, error) {
	return s.webappProvider.FindByPolicyId(id, ctx)
}

func (s *Service) List(ctx context.Context) ([]PolicyListItem, error) {
	policies, err := s.repository.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]PolicyListItem, 0, len(policies))

	policyIds := make([]primitive.ObjectID, len(policies))
	for i := range policies {
		policyIds[i] = policies[i].ID
	}

	webappsMap, err := s.webappProvider.FindByPolicyIDs(policyIds, ctx)
	if err != nil {
		return nil, err
	}

	for _, policy := range policies {
		items = append(items, PolicyListItem{
			ID:      policy.ID,
			Name:    policy.Name,
			WL:      policy.WL,
			Webapps: webappsMap[policy.ID],
		})
	}
	return items, nil
}

func (s *Service) Delete(ctx context.Context, entity *Policy) error {
	return s.repository.Delete(ctx, entity)
}

func (s *Service) Update(ctx context.Context, entity *Policy) error {
	return s.repository.Update(ctx, entity)
}

func (s *Service) CanDeletePolicy(ctx context.Context, id primitive.ObjectID) error {
	webapps, err := s.webappProvider.FindByPolicyId(id, ctx)
	if err != nil {
		return err
	}
	if len(webapps) > 0 {
		return &PolicyInUseError{Webapps: webapps}
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
