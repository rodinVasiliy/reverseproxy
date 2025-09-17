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
		if err != nil || record == nil || record.Country.IsoCode != "RU" {
			fmt.Println("request will be blocked by GEO")
			return false
		}
		return true
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
	// to do мб базу надо в другом месте закрывать, т.к. тут она может сразу закрыться

	return []Rule{
		{
			name:     "GEO",
			actions:  actions,
			ruleFunc: GEO(),
		},
	}
}
