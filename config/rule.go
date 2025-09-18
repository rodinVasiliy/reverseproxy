package config

import (
	"fmt"
	"log"
	"net/http"

	"github.com/oschwald/geoip2-golang"
)

type Rule struct {
	name     string
	actions  []Action
	ruleFunc func(r *http.Request, policy *Policy) bool
}

var geoDB *geoip2.Reader

func GEO() func(r *http.Request, policy *Policy) bool {
	return func(r *http.Request, policy *Policy) bool {
		ip := getIpFromRequest(r)
		record, err := geoDB.Country(ip)
		// блокируем РФ, остальное - пропускаем
		// поделить на секции чтобы видеть когда ошибка, а когда запись nil и когда уже норм отработала база.

		if err != nil {
			log.Println("error reading geo base for request. Request will be blocked by GEO")
			fmt.Println("error reading geo base for request. Request will be blocked by GEO")
			return false
		}

		if record == nil {
			log.Println("record is nil. Request will be blocked by GEO")
			fmt.Println("record is nil. Request will be blocked by GEO")
			return false
		}

		if record.Country.IsoCode != "RU" {
			fmt.Println("Geo is not Russia. Request will be not blocked by GEO")
			log.Println("Geo is not Russia. Request will be not blocked by GEO")
			return false
		}

		fmt.Println("Geo is Russia, Request will be blocked by GEO")
		log.Println("Geo is Russia, Request will be blocked by GEO")
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
