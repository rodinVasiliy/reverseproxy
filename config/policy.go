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

	var actions []Action

	ipStr, _, _ := net.SplitHostPort(r.RemoteAddr)
	ip := net.ParseIP(ipStr)
	if p.wl != nil {
		if ok := checkInList(p.wl, ip); ok {
			return nil
		}
	}
	for _, rule := range p.rules {
		if ok := rule.ruleFunc(r, p); ok {
			actions = append(actions, rule.actions...)
		}
	}

	return actions
}

func DefaultPolicy() Policy {
	rules := InitRules()
	var p = Policy{nil, rules}
	return p
}
