package checkrequest

import (
	"fmt"
	"net"
	"net/http"
	act "reverseproxy/model/action"
	parsed_request "reverseproxy/model/parsed_request"
	service "reverseproxy/service"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// проверяет, что ip есть в списке(может использоваться и WL и BL)
func checkInList(ip net.IP, list []string) bool {
	for i := range list {
		_, net, _ := net.ParseCIDR(list[i])
		if net.Contains(ip) {
			return true
		}
	}
	return false
}

// проверяет есть ли в actions блок
// логирует запрос если есть лог - для этого и нужно ruleName в параметрах
// добавляет список уникальных actions чтобы потом их выполнил другой метод
func checkActions(actionIds []primitive.ObjectID, as *service.ActionService, pr *parsed_request.ParsedRequest, ruleName string,
	uniqueActions map[string]struct{}) bool {
	isBlock := false
	for i := range actionIds {
		actionDoc, _ := as.FindById(actionIds[i])
		action := act.ActionByName(actionDoc.Name)
		switch action.ActionType {
		case act.ActionBlock:
			isBlock = true
		case act.ActionLog:
			actionParams := &act.ActionParams{Rule: ruleName, PR: pr}
			action.Do(actionParams)
		default:
			uniqueActions[ruleName] = struct{}{}
		}
	}
	return isBlock
}

// проверка, нужно ли блокировать запрос, проходимся по всем правилам из политики
func IsBlock(r *http.Request, ws *service.WebAppService, ps *service.PolicyService,
	rs *service.RuleService, as *service.ActionService) (bool, error) {
	host := r.Host
	if h, _, err := net.SplitHostPort(r.Host); err == nil {
		host = h
	}
	webapp, err := ws.GetWebAppForHost(host)
	if err != nil {
		return false, fmt.Errorf("failed to find web app for host %s %s", host, err)
	}
	policy, err := ps.FindById(webapp.PolicyId)
	if err != nil {
		return false, fmt.Errorf("failed to find policy %s", err)
	}
	parsedRequest := parsed_request.ParseRequest(r)
	if checkInList(parsedRequest.IP, policy.WL) {
		return false, nil
	}
	isBlock := false
	uniqueActions := map[string]struct{}{}
	// TODO - переделать логику
	for _, ruleRef := range policy.Rules {
		rule, err := rs.FindById(ruleRef.RuleID)
		if err != nil {
			return false, fmt.Errorf("failed to find rule %s by id %s", rule.Name, err)
		}
		if rule.Match(parsedRequest.ToMap()) {
			// точно ли nil?
			if ruleRef.Actions == nil {
				// это если у правила используются default actions(нет переопределения для actions)
				actions := rule.Actions
				// переименовать функцию
				if ok := checkActions(actions, as, parsedRequest, rule.Name, uniqueActions); ok {
					isBlock = true
				}
			} else {
				// это если у правила переопределен список actions
				if ok := checkActions(ruleRef.Actions, as, parsedRequest, rule.Name, uniqueActions); ok {
					isBlock = true
				}
			}
		}
	}
	for actionDoc := range uniqueActions {
		action := act.ActionByName(actionDoc)
		actionParams := act.ActionParams{Rule: "", PR: parsedRequest}
		action.Do(&actionParams)
	}
	return isBlock, nil
}
