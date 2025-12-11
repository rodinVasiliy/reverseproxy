package parsedrequest

import (
	"net"
	"net/http"
	geo "reverseproxy/config/geo"
	utils "reverseproxy/utils"
)

type ParsedRequest struct {
	IP          net.IP
	Host        string
	Path        string
	Uri         string
	Method      string
	UA          string
	CountryCode string
	Cookies     []*http.Cookie
}

func ParseRequest(r *http.Request) *ParsedRequest {
	var countryCode string
	ip := utils.GetIpFromRequest(r)

	countryCode = geo.GetGeoCode(ip)

	rp := ParsedRequest{
		IP:          ip,
		Host:        r.Host,
		Path:        r.URL.Path,
		UA:          r.Header.Get("User-Agent"),
		Cookies:     r.Cookies(),
		Uri:         r.RequestURI,
		Method:      r.Method,
		CountryCode: countryCode,
	}
	return &rp
}

func (rp *ParsedRequest) ToMap() map[string]string {
	result := map[string]string{
		IP:           rp.IP.String(),
		HOST:         rp.Host,
		PATH:         rp.Path,
		URI:          rp.Uri,
		rp.Method:    rp.Method,
		UA:           rp.UA,
		COUNTRY_CODE: rp.CountryCode,
	}
	for _, cookie := range rp.Cookies {
		name := "cookie:" + cookie.Name
		result[name] = cookie.Value
	}
	return result
}

func GetCookie(params map[string]string, name string) (string, bool) {
	key := "cookie:" + name
	val, ok := params[key]
	return val, ok
}

var IP = "ip"
var HOST = "host"
var PATH = "path"
var URI = "uri"
var METHOD = "method"
var UA = "ua"
var COUNTRY_CODE = "countryCode"
