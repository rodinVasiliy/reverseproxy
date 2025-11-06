package service

import (
	"net/http"
	config "reverseproxy/config/mongo_config"
	"reverseproxy/model/action"
	parsedrequest "reverseproxy/model/parsed_request"
	policy "reverseproxy/model/policy"
)

type PolicyService struct {
	deps        *config.MongoDeps
	ruleService *RuleService
}

func NewPolicyService(deps *config.MongoDeps, ruleService *RuleService) *PolicyService {
	return &PolicyService{deps: deps, ruleService: ruleService}
}

// ////////////////////// INIT SECTION
func (ps *PolicyService) getDefaultPolicy() policy.Policy {
	rules := ps.ruleService.FindAllRules()

	// добавляем к правилам дефолтные actions
	var policyRuleRef []policy.PolicyRuleRef
	for _, rule := range rules {
		policyRuleRef = append(policyRuleRef, policy.PolicyRuleRef{RuleID: rule.ID})
	}
	// wl пока будет пустым
	defaultPolicy := policy.Policy{WL: nil, Rules: policyRuleRef}
	return defaultPolicy
}

func (ps *PolicyService) LoadDefaultPolicyToDB() {
	mongoConfig := ps.deps.Config
	client := ps.deps.Client

	db := mongoConfig.Database
	collection := client.Database(db).Collection(POLICY_COLLECTION)
	ctx, cancel := ps.deps.Ctx()
	defer cancel()

	policy := ps.getDefaultPolicy()
	collection.InsertOne(ctx, policy)
}

//////////////////////// END INIT SECTION

// проверяем, нужно ли блокировать реквест
// + проходим все actions, которые вернули правила.
func (ps *PolicyService) IsBlockedByPolicy(r *http.Request) bool {

	parsedRequest := parsedrequest.ParseRequest(r)

	var uniqueActions = make(map[string]action.Action)

	// ip := parsedRequest.ip

	// // Белый список: если IP в whitelist → сразу пропускаем
	// if p.wl != nil {
	// 	if ok := checkInList(p.wl, ip); ok {
	// 		log.Printf("request will be passed by policy. IP %s in wl", ip.String())
	// 		return true
	// 	}
	// }

	// // проверка на наличие в BL, выглядит пока так себе, возможно надо будет менять структуру и логику всего реверс прокси.
	// ok, err := p.bl.Exists(ip.String())
	// if err != nil {
	// 	log.Printf("failed to check ip in BL %s", err)
	// }
	// if ok {
	// 	log.Printf("ip %s in BL, request will be blocked", ip.String())
	// 	return true
	// }

	var blockRequest = false
	// Пробегаем по правилам
	// попробовать распараллелить, получив например количество правил и распределив по потокам...
	requestMap := parsedRequest.ToMap()
	// получать из базы список ActionDoc, по ним получать настоящие Actions

	for _, ruleId := range p.Rules {

	}

	// выполняем уникальные экшены
	ap := ActionParams{rp: parsedRequest}
	for _, act := range uniqueActions {
		act.Do(&ap)
	}

	return blockRequest
}
