package parsedrequest

import (
	"net"
	"net/http"
	geo "reverseproxy/internal/infrastructure/config/geo"
	"reverseproxy/internal/utils"
)

type ParsedRequest struct {
	IP          net.IP
	Host        string
	Path        string
	Method      string
	UA          string
	CountryCode string
	Cookies     []*http.Cookie
}

func NewParsedRequest(r *http.Request) *ParsedRequest {
	var countryCode string
	ip := utils.GetIpFromRequest(r)

	countryCode = geo.GetGeoCode(ip)

	rp := ParsedRequest{
		IP:          ip,
		Host:        r.Host,
		Path:        r.URL.Path,
		UA:          r.Header.Get("User-Agent"),
		Cookies:     r.Cookies(),
		Method:      r.Method,
		CountryCode: countryCode,
	}
	return &rp
}

func (rp *ParsedRequest) ToMap() map[string]string {
	result := map[string]string{
		IP:          rp.IP.String(),
		HOST:        rp.Host,
		PATH:        rp.Path,
		METHOD:      rp.Method,
		UA:          rp.UA,
		CountryCode: rp.CountryCode,
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

var IP = "IP"
var HOST = "Host"
var PATH = "Path"
var METHOD = "Method"
var UA = "UA"
var CountryCode = "CountryCode"
