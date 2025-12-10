package checkrequest

import (
	"fmt"
	"log"
	"net"
	"net/http"
	action "reverseproxy/internal/model/action"
	parsed_request "reverseproxy/internal/model/parsed_request"
	policy "reverseproxy/internal/model/policy"
	rule "reverseproxy/internal/model/rule"
	webapp "reverseproxy/internal/model/webapp"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// возвращает true, если ip есть в списке подсетей list (может использоваться и WL и BL)
func checkInList(ip net.IP, list []string) bool {
	for i := range list {
		_, net, err := net.ParseCIDR(list[i])
		if err != nil {
			log.Printf("failed to parse cidr %s %s", list[i], err)
		}
		if net.Contains(ip) {
			return true
		}
	}
	return false
}

// проверка, нужно ли блокировать запрос, проходимся по всем правилам из политики
func IsBlock(r *http.Request, ws *webapp.Service, ps *policy.Service,
	rs *rule.Service, as *action.Service, executor *action.Executor) (bool, error) {
	host := r.Host
	if h, _, err := net.SplitHostPort(r.Host); err == nil {
		host = h
	}
	webapp, err := ws.GetWebAppForHost(r.Context(), host)
	if err != nil {
		return false, fmt.Errorf("failed to find web app for host %s %s", host, err)
	}
	policy, err := ps.FindById(r.Context(), webapp.PolicyId)
	if err != nil {
		return false, fmt.Errorf("failed to find policy %s", err)
	}
	parsedRequest := parsed_request.ParseRequest(r)
	// проверяем наличие объекта в WL политики
	if checkInList(parsedRequest.IP, policy.WL) {
		return false, nil
	}

	shouldBlock := false
	for _, ruleRef := range policy.Rules {
		rule, err := rs.FindById(r.Context(), ruleRef.RuleID)
		if err != nil {
			return false, fmt.Errorf("failed to find rule %s by id %s", rule.Name, err)
		}
		if rule.Match(parsedRequest.ToMap()) {
			// точно ли nil?
			var actionIDs []primitive.ObjectID
			if ruleRef.Actions == nil {
				// это если у правила используются default actions(нет переопределения для actions)
				actionIDs = rule.Actions
			} else {
				// это если у правила переопределен список actions
				actionIDs = ruleRef.Actions
			}
			actionDocs, err := as.FindByIds(r.Context(), actionIDs)
			if err != nil {
				return false, fmt.Errorf("failed to find actions by ids: %w", err)
			}
			for _, act := range actionDocs {
				if ok := executor.Execute(act.Name, &action.ActionParams{Rule: rule.Name, PR: parsedRequest}); ok {
					shouldBlock = true
				}
			}
		}
	}
	return shouldBlock, nil
}
