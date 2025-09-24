package config

import (
	"fmt"
	"log"
	"net"
	"net/http"
)

type Policy struct {
	wl    []*net.IPNet
	rules []Rule
	bl    *BL
}

// проверяем, нужно ли блокировать реквест
// + проходим все actions, которые вернули правила.
func (p *Policy) IsBlockedByPolicy(r *http.Request) bool {

	parsedRequest := ParseRequest(r)

	var uniqueActions = make(map[string]Action)

	ip := parsedRequest.ip

	// Белый список: если IP в whitelist → сразу пропускаем
	if p.wl != nil {
		if ok := checkInList(p.wl, ip); ok {
			log.Printf("request will be passed by policy. IP %s in wl", ip.String())
			return true
		}
	}

	// проверка на наличие в BL, выглядит пока так себе, возможно надо будет менять структуру и логику всего реверс прокси.
	ok, err := p.bl.Exists(ip.String())
	if err != nil {
		log.Printf("failed to check ip in BL %s", err)
	}
	if ok {
		log.Printf("ip %s in BL, request will be blocked", ip.String())
		return true
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

func DefaultPolicy(bl *BL) (*Policy, error) {
	rules := InitRules()

	// WL(пока что тестовый)
	var wl []*net.IPNet
	wlCidr := "109.169.245.0/24"
	_, ipnet, err := net.ParseCIDR(wlCidr)
	if err != nil {
		return nil, fmt.Errorf("error parsing WL")
	}
	wl = append(wl, ipnet)

	var p = Policy{wl, rules, bl}
	return &p, nil
}
