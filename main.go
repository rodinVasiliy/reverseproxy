package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"reverseproxy/config"
)

func main() {
	fmt.Println("reverse proxy ...")
	fmt.Println("Config Initialization started...")
	config, err := config.InitConfig()
	if err != nil {
		fmt.Printf("failed to read config %s", err)
	}
	fmt.Println("Config successfully loaded")

	// порт который использует прокси сервер, мы будем передавать его в заголовок, просто для инфо
	port := flag.Int("port", 9001, "Port for proxy serv")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("X-Proxy-Port", fmt.Sprintf("%d", *port))
		log.Printf("Proxy request %s %s via port %d", r.Method, r.URL.Path, *port)
		proxy := getProxyForRequest(r, config)
		proxy.ServeHTTP(w, r)
	})

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Starting proxy on %s\n", addr)

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func getProxyForRequest(r *http.Request, cfg *config.Config) *httputil.ReverseProxy {
	host := r.URL.Host
	proxy := cfg.GetReverseProxyForHost(host)
	return proxy
}
