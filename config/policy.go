package config

import (
	"net"
	"net/http"
)

type Policy struct {
	wl    []*net.IPNet
	rules []Rule
}

func (p *Policy) checkRequest(w http.ResponseWriter, r *http.Request) []Action {

	var uniqueActions = make(map[string]Action)

	ipStr, _, _ := net.SplitHostPort(r.RemoteAddr)
	ip := net.ParseIP(ipStr)

	// Белый список: если IP в whitelist → сразу пропускаем
	if p.wl != nil {
		if ok := checkInList(p.wl, ip); ok {
			return nil
		}
	}

	// Пробегаем по правилам
	for _, rule := range p.rules {
		if ok := rule.ruleFunc(r, p); ok {
			for _, act := range rule.actions {
				// используем имя/идентификатор действия как ключ
				uniqueActions[act.Name()] = act
			}
		}
	}

	// Конвертируем map → slice
	var actions []Action
	for _, act := range uniqueActions {
		actions = append(actions, act)
	}

	return actions
}

func DefaultPolicy() Policy {
	rules := InitRules()
	var p = Policy{nil, rules}
	return p
}
