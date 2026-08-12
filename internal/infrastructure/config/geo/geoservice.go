package geo

import (
	"fmt"
	"net"

	"github.com/oschwald/geoip2-golang"
)

// Секция с ГЕО
var geoDB *geoip2.Reader
var geoPath = "internal/infrastructure/config/geo/geo_config/dbip-country-lite-2025-09.mmdb"

func CloseGeoDB() {
	if geoDB != nil {
		geoDB.Close()
		fmt.Println("geo base closed")
	}
}

func InitGeo() error {
	var err error
	geoDB, err = geoip2.Open(geoPath)
	if err != nil {
		return err
	} else {
		fmt.Println("Geo base successfully loaded")
		return nil
	}
}

func GetGeoCode(ip net.IP) string {
	record, err := geoDB.Country(ip)
	var countryCode string
	if err != nil {
		countryCode = ""
	} else {
		countryCode = record.Country.IsoCode
	}
	return countryCode
}
