package config

import (
	"fmt"
	"net"
	"net/http"
)

type Policy struct {
	wl    []*net.IPNet
	rules []Rule
}

// проверяем, нужно ли блокировать реквест
// + проходим все actions, которые вернули правила.
func (p *Policy) CheckRequest(r *http.Request) bool {

	parsedRequest := ParseRequest(r)

	var uniqueActions = make(map[string]Action)

	ip := parsedRequest.ip

	// Белый список: если IP в whitelist → сразу пропускаем
	if p.wl != nil {
		if ok := checkInList(p.wl, ip); ok {
			return true
		}
	}

	// Пробегаем по правилам
	// попробовать распараллелить, получив например количество правил и распределив по потокам...
	for _, rule := range p.rules {
		if ok := rule.ruleFunc(parsedRequest, p); ok {
			for _, act := range rule.actions {
				// логируем сразу(так проще, пока не придумал как это еще делать)
				if act.Name() == "Log to DB" {
					fmt.Println("will add raw to db")
					act.DoAction(&ActionParams{
						rule: rule.name,
						rp:   parsedRequest,
					})
				} else { // если не логируем - кладем в список функций которые будем выполнять,
					// нужно чтобы каждая выполнилась 1 раз, без повторений
					uniqueActions[act.Name()] = act
				}
			}
		}
	}

	var blockRequest = false

	// если был блок - значит говорим, что запрос нужно блокировать
	// убираем этот Action, так как он просто сигнал для нас
	if _, ok := uniqueActions["Block Request"]; ok {
		blockRequest = true
		delete(uniqueActions, "Block Request")
	}

	ap := ActionParams{
		rule: "",
		rp:   parsedRequest,
	}
	for _, act := range uniqueActions {
		act.DoAction(&ap)
	}

	return blockRequest
}

func DefaultPolicy() Policy {
	rules := InitRules()
	var p = Policy{nil, rules}
	return p
}
