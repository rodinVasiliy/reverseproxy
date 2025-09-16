package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	cfg "reverseproxy/config"
)

// TO DO протестировать
func main() {
	fmt.Println("reverse proxy ...")
	fmt.Println("Config Initialization started...")
	config, err := cfg.InitConfig()
	if err != nil {
		fmt.Printf("failed to read config %s", err)
	}
	fmt.Println("Config successfully loaded")

	// порт который использует прокси сервер, мы будем передавать его в заголовок, просто для инфо
	port := flag.Int("port", 9001, "Port for proxy serv")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sendThroughPolicy(w, r, config)

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

	cfg.CloseGeoDB()
}

func getProxyForRequest(r *http.Request, cfg *cfg.Config) *httputil.ReverseProxy {
	host := r.URL.Host
	proxy := cfg.GetReverseProxyForHost(host)
	return proxy
}

func sendThroughPolicy(w http.ResponseWriter, r *http.Request, cfg *cfg.Config) {
	actions := cfg.CheckRequest(w, r)
	for _, act := range actions {
		act.DoAction(w, r)
	}
}
