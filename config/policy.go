package config

import (
	"net"
	"net/http"
)

type Policy struct {
	wl    []*net.IPNet
	bl    []*net.IPNet
	rules []Rule
}

func (p *Policy) checkRequest(r *http.Request) []Action {

	var actions []Action

	// check WL

	// check ip in BL

	// go for all rules in []rules
	// add actions to result
	// return Actions

	return actions
}
