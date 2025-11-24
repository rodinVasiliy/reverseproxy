package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	check_request "reverseproxy/check_request"
	log_config "reverseproxy/config/log"
	config "reverseproxy/config/mongo_config"
	"reverseproxy/geo"
	wafconfig "reverseproxy/model/waf_config"
	service "reverseproxy/service"
	initialization "reverseproxy/service/initialization"
	"strconv"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func getPort() (int, error) {
	portStr := os.Getenv("PROXY_PORT")
	if portStr == "" {
		return -1, fmt.Errorf("please executing program don't forget enter proxy port")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return -1, fmt.Errorf("failed to parse proxy port %s %w", portStr, err)
	}
	return port, nil
}

func getInItFlag() bool {
	return os.Getenv("INIT") == "1"
}

// TO DO протестировать
func main() {

	fmt.Println("reverse proxy ...")

	// порт который использует прокси сервер, мы будем передавать его в заголовок, просто для инфо
	port, err := getPort()
	if err != nil {
		fmt.Printf("failed to get proxy port %s", err)
		return
	}

	// тут идет настройка лог файла, в котором будут отображаться ошибки
	logConfig, err := log_config.NewLogConfig(port)
	if err != nil {
		fmt.Printf("failed to open log file %s", err)
		return
	}

	// загружаем гео базу
	fmt.Println("loading geo base from file ...")
	err = geo.InitGeo()
	if err != nil {
		fmt.Printf("failed to in it geo base %s", err)
		closeAll(logConfig)
		return
	}

	fmt.Println("Getting mongo db dependencies ...")
	// подключаемся к монгодб
	mongoDeps, err := config.NewMongoDeps()
	if err != nil {
		fmt.Printf("failed to get mongo deps %s", err)
		closeAll(logConfig)
		return
	}

	fmt.Println("Config Initialization started...")
	actionService := service.NewActionService(mongoDeps)
	ruleService := service.NewRuleService(mongoDeps, actionService)
	policyService := service.NewPolicyService(mongoDeps, ruleService)
	sslService := service.NewSSLConfigurationService(mongoDeps)
	webAppService := service.NewWebAppService(mongoDeps)

	if getInItFlag() {
		fmt.Println("Initialization database ...")
		err = initialization.InItDB(policyService, actionService, ruleService)
		if err != nil {
			fmt.Printf("failed to in it db %s", err)
			closeAll(logConfig)
			return
		}
		err = initialization.NewTestWebApp(policyService, sslService, webAppService)
		if err != nil {
			fmt.Printf("failed to add test webapp %s", err)
			closeAll(logConfig)
			return
		}
	} else {
		fmt.Println("Init db not required")
	}

	// конфиг нужен, чтобы для хоста выдавать httputil.ReverseProxy
	wafConfig, err := wafconfig.NewWafConfig(webAppService)
	if err != nil {
		fmt.Printf("failed to load waf config %s", err)
		closeAll(logConfig)
		return
	}
	fmt.Println("Waf Config successfully loaded")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Proxy request %s %s via port %d", r.Method, r.URL.Path, port)

		// проверяем, нужно ли блокировать запрос
		if check_request.IsBlock(r, webAppService, policyService, ruleService, actionService) {
			http.Error(w, "Access Denied", http.StatusForbidden)
			return
		}

		r.Header.Set("X-Proxy-Port", fmt.Sprintf("%d", port))

		webApp, err := webAppService.GetWebAppForHost(r.Host)
		if err != nil {
			fmt.Printf("failed to get web app for host %s, %s", r.Host, err)
		}
		proxy := wafConfig.GetProxyForWebApp(webApp)
		// TODO что делать если ничего не нашлось
		log.Printf("Forward request %s %s to upstream", r.Method, r.URL.Path)
		proxy.ServeHTTP(w, r)
	})

	// слушаем только с nginx
	addr := fmt.Sprintf("127.0.0.1:%d", port)

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

	closeAll(logConfig)

	log.Println("Server stopped")
}

// закрываем нужные ресурсы - файл для лога и гео базу
func closeAll(logConfig *log_config.LogConfig) {
	logConfig.CloseLogFile()
	geo.CloseGeoDB()
}
