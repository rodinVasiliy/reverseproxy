package config

import (
	"fmt"
	"log"

	"github.com/oschwald/geoip2-golang"
)

type Rule struct {
	name     string
	actions  []Action                                     // то, что висит на самом правиле, от этого зависит, что будет происходить, если запрос попадет под правило.
	ruleFunc func(rp *RequestParams, policy *Policy) bool // true - если запрос попадает под правило. false - иначе.
}

var geoDB *geoip2.Reader

func GEO() func(rp *RequestParams, policy *Policy) bool {
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

		if record.Country.IsoCode == "RU" {
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
			name:     "GEO",
			actions:  actions,
			ruleFunc: GEO(),
		},
	}
}
