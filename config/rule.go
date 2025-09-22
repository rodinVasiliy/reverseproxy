package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/oschwald/geoip2-golang"
)

type Rule struct {
	name     string
	actions  []Action                                     // то, что висит на самом правиле, от этого зависит, что будет происходить, если запрос попадет под правило.
	ruleFunc func(rp *RequestParams, policy *Policy) bool // true - если запрос попадает под правило. false - иначе.
}

// перенести в config
var geoDB *geoip2.Reader

func blockByGeoIp() func(rp *RequestParams, policy *Policy) bool {
	return func(rp *RequestParams, policy *Policy) bool {
		ip := rp.ip
		record, err := geoDB.Country(ip)

		// блокируем(пока только логируем) РФ, остальное - пропускаем
		// поделить на секции чтобы видеть когда ошибка, а когда запись nil и когда уже норм отработала база.

		if err != nil {
			fmt.Println("error reading geo base for request. Request will be blocked by GEO")
			return true
		}

		if record == nil {
			fmt.Println("record is nil. Request will be blocked by GEO")
			return true
		}

		// блокируем если РФ(пока только логируем)
		if record.Country.IsoCode == "RU" {
			return true
		}

		return false
	}
}

func blockByUserAgent() func(rp *RequestParams, policy *Policy) bool {
	return func(rp *RequestParams, policy *Policy) bool {
		ua := rp.ua
		// блокируем если UA нет
		if ua == "" {
			return true
		}

		// блокируем если UA не содержит Mozilla/5.0(для нас легитим только браузерные UA)
		ua = strings.ToLower(ua)
		if ok := strings.Contains(ua, "Mozilla/5.0"); !ok {
			return true
		}
		return false
	}
}

func CloseGeoDB() {
	if geoDB != nil {
		geoDB.Close()
		fmt.Println("geo base closed")
	}
}

func InitRules() []Rule {
	var err error
	geoDB, err = geoip2.Open("config/dbip-country-lite-2025-09.mmdb")
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("Geo base successfully loaded")
	}

	actions := InitActions()

	return []Rule{
		{
			name:     "block by geo ip",
			actions:  actions,
			ruleFunc: blockByGeoIp(),
		},
		{
			name:     "block by user agent",
			actions:  actions,
			ruleFunc: blockByUserAgent(),
		},
	}
}
