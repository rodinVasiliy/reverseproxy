package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	cfg "reverseproxy/config"
	"syscall"
	"time"
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

	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Starting proxy on %s\n", addr)
		if err := http.ListenAndServe(addr, handler); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	<-stop
	log.Println("Shitting down the server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Закрываем сервер
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	cfg.CloseGeoDB()

	log.Println("Server stopped")
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
