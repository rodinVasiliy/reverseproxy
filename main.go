package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/oschwald/geoip2-golang"
)

var geoDB *geoip2.Reader

func main() {
	fmt.Println("reverse proxy ...")

	fmt.Println("getting geo base ...")
	var err error
	geoDB, err = geoip2.Open("GeoLite2-Country.mmdb")
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("Geo base successfully loaded")
	}
	defer geoDB.Close()

	port := flag.Int("port", 9001, "Port for proxy serv")
	upstream := flag.String("upstream", "http://localhost:9091", "Upstream server address")

	flag.Parse()
	target, err := url.Parse(*upstream)
	if err != nil {
		log.Fatalf("Failed to parse upstream: %v", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(target)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ipStr, _, _ := net.SplitHostPort(r.RemoteAddr)
		ip := net.ParseIP(ipStr)
		record, err := geoDB.Country(ip)
		if err != nil || record == nil || record.Country.IsoCode != "RU" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		r.Header.Set("X-Proxy-Port", fmt.Sprintf("%d", *port))
		log.Printf("Proxy request %s %s via port %d", r.Method, r.URL.Path, *port)
		proxy.ServeHTTP(w, r)
	})

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Starting proxy on %s, forwarding to %s", addr, *upstream)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
