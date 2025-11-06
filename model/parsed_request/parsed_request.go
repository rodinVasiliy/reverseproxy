package parsedrequest

import (
	"net"
	"net/http"
	geo "reverseproxy/geo"
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
	ip := getIpFromRequest(r)

	countryCode = geo.GetGeoCode(ip)

	rp := ParsedRequest{
		IP:          ip,
		Host:        r.Host,
		Path:        r.URL.RawPath,
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
		"ip":          rp.IP.String(),
		"host":        rp.Host,
		"path":        rp.Path,
		"uri":         rp.Uri,
		"method":      rp.Method,
		"ua":          rp.UA,
		"countryCode": rp.CountryCode,
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

func getIpFromRequest(r *http.Request) net.IP {
	xrip := r.Header.Get("X-Real-IP")
	return net.ParseIP(xrip)
}
