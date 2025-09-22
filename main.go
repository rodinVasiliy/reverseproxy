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

	_ "github.com/mattn/go-sqlite3"
)

// TO DO протестировать
func main() {

	fmt.Println("reverse proxy ...")

	// порт который использует прокси сервер, мы будем передавать его в заголовок, просто для инфо
	// TODO - сделать корректный парсинг, чтобы выдавало ошибку, если не получится считать число
	port := flag.Int("port", 9001, "Port for proxy serv")
	flag.Parse()

	fmt.Println("Config Initialization started...")
	config, err := cfg.InitConfig(*port)
	if err != nil {
		fmt.Printf("failed to read config %s", err)
		return
	}
	fmt.Println("Config successfully loaded")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Proxy request %s %s via port %d", r.Method, r.URL.Path, *port)

		if !sendThroughPolicy(r, config) {
			http.Error(w, "Access Denied", http.StatusForbidden)
		}

		r.Header.Set("X-Proxy-Port", fmt.Sprintf("%d", *port))

		proxy := getProxyForRequest(r, config)
		log.Printf("Forward request %s %s to upstream", r.Method, r.URL.Path)
		proxy.ServeHTTP(w, r)
	})

	// слушаем только с nginx
	addr := fmt.Sprintf("127.0.0.1:%d", *port)

	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		fmt.Printf("Starting proxy on %s\n", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Server failed: %v", err)
			return
		}
	}()

	<-stop
	fmt.Println("Shutting down the server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Закрываем сервер
	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("Server forced to shutdown: %v", err)
	}

	cfg.CloseGeoDB()
	config.CloseLogFile()

	log.Println("Server stopped")
}

func getProxyForRequest(r *http.Request, cfg *cfg.Config) *httputil.ReverseProxy {
	host := r.Host
	proxy := cfg.GetReverseProxyForHost(host)
	return proxy
}

func sendThroughPolicy(r *http.Request, cfg *cfg.Config) bool {
	host := r.Host
	policy := cfg.GetPolicyForHost(host)
	return policy.CheckRequest(r)
}
