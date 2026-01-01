package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	check_request "reverseproxy/check_request"
	bl "reverseproxy/config/bl"
	"reverseproxy/config/geo"
	log_config "reverseproxy/config/log"
	config "reverseproxy/config/mongo_config"
	"reverseproxy/internal/http/handler"
	"reverseproxy/internal/http/middleware"
	action "reverseproxy/internal/model/action"
	policy "reverseproxy/internal/model/policy"
	rule "reverseproxy/internal/model/rule"
	ssl "reverseproxy/internal/model/ssl"
	wafconfig "reverseproxy/internal/model/waf_config"
	webapp "reverseproxy/internal/model/webapp"
	repository "reverseproxy/internal/repository"

	initialization "reverseproxy/initialization"
	utils "reverseproxy/utils"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
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

// TO DO разобраться с логами
func main() {

	// fmt в консольку для информации - вся инфа до такого как все запустилось
	// log - для ошибок при работе с базой и прочим
	fmt.Println("reverse proxy ...")

	// порт который использует прокси сервер, мы будем передавать его в заголовок, просто для инфо
	port, err := getPort()
	if err != nil {
		fmt.Printf("failed to get proxy port %s", err)
		return
	}

	errorLogFileName := filepath.Join("log", fmt.Sprintf("error_%d.log", port))
	accessLogFileName := filepath.Join("log", fmt.Sprintf("access_%d.log", port))

	// тут идет настройка лог файла, в котором будут отображаться ошибки
	errorLogConfig, err := log_config.NewLogConfig(errorLogFileName)
	if err != nil {
		fmt.Printf("failed to open log file %s :%s", errorLogFileName, err)
		return
	}
	// пока что все ошибки будут логироваться в error log
	log.SetOutput(errorLogConfig.File())
	accessLogConfig, err := log_config.NewLogConfig(accessLogFileName)
	if err != nil {
		fmt.Printf("failed to open log file %s :%s", accessLogFileName, err)
		closeAll(nil, errorLogConfig, nil)
		return
	}
	accessLogger := log.New(accessLogConfig.File(), "", log.LstdFlags|log.Lmicroseconds)

	var blackList *bl.RedisBL
	blAddr := "localhost:9999"
	blackList, err = bl.NewRedisBL(blAddr, "", 0)
	if err != nil {
		fmt.Printf("failed to in it bl %s", err)
		closeAll(blackList, errorLogConfig, accessLogConfig)
		return
	}

	// загружаем гео базу
	fmt.Println("loading geo base from file ...")
	err = geo.InitGeo()
	if err != nil {
		fmt.Printf("failed to in it geo base %s", err)
		closeAll(blackList, errorLogConfig, accessLogConfig)
		return
	}

	fmt.Println("Getting mongo db dependencies ...")
	// подключаемся к монгодб
	mongoDeps, err := config.NewMongoDeps()
	if err != nil {
		fmt.Printf("failed to get mongo deps %s", err)
		closeAll(blackList, errorLogConfig, accessLogConfig)
		return
	}

	fmt.Println("Config Initialization started...")
	actionRepository := repository.NewMongoRepositoy[action.ActionDoc](mongoDeps.Client, repository.DB_NAME, repository.ACTION_COLLECTION)
	actionService := action.NewService(actionRepository)
	actionRegistry := action.NewActionRegistry(accessLogger, blackList)
	actionExecutor := action.NewExecutor(actionRegistry)

	ruleRepository := repository.NewMongoRepositoy[rule.Rule](mongoDeps.Client, repository.DB_NAME, repository.RULE_COLLECTION)
	ruleService := rule.NewService(ruleRepository)

	policyRepository := repository.NewMongoRepositoy[policy.Policy](mongoDeps.Client, repository.DB_NAME, repository.POLICY_COLLECTION)
	policyService := policy.NewService(policyRepository)

	sslRepository := repository.NewMongoRepositoy[ssl.SSLConfiguration](mongoDeps.Client, repository.DB_NAME, repository.SSL_COLLECTION)
	sslService := ssl.NewService(sslRepository)

	webappRepository := repository.NewMongoRepositoy[webapp.WebApp](mongoDeps.Client, repository.DB_NAME, repository.WEBAPP_COLLECTION)
	webAppService := webapp.NewService(webappRepository, sslService)

	// API
	r := gin.Default()
	r.Use(middleware.CORS())
	api := r.Group("/api")
	handler.RegisterActionRoutes(api, actionService)
	handler.RegisterPolicyRoutes(api, policyService)
	handler.RegisterSSLRoutes(api, sslService)
	handler.RegisterWebAppRoutes(api, webAppService)

	if getInItFlag() {
		fmt.Println("Initialization database ...")
		utils.DropAllCollections(mongoDeps)
		err = initialization.InItDB(policyService, actionService, ruleService)
		if err != nil {
			fmt.Printf("failed to in it db %s", err)
			closeAll(blackList, errorLogConfig, accessLogConfig)
			return
		}
		err = initialization.NewTestWebApp(policyService, sslService, webAppService)
		if err != nil {
			fmt.Printf("failed to add test webapp %s", err)
			closeAll(blackList, errorLogConfig, accessLogConfig)
			return
		}
	} else {
		fmt.Println("Init db not required")
	}

	// конфиг нужен, чтобы для хоста выдавать httputil.ReverseProxy
	wafConfig, err := wafconfig.NewWafConfig(webAppService)
	if err != nil {
		fmt.Printf("failed to load waf config %s", err)
		closeAll(blackList, errorLogConfig, accessLogConfig)
		return
	}
	fmt.Println("Waf Config successfully loaded")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Proxy request %s %s %s via port %d\n", r.Host, r.Method, r.URL.Path, port)
		host := r.Host
		ip := utils.GetIpFromRequest(r)
		ok, err := blackList.Exists(ip.String())
		if err != nil {
			fmt.Printf("failed to check request in BL %s", err)
		}
		if ok {
			fmt.Println("access will be denied")
			http.Error(w, "Access Denied", http.StatusForbidden)
			return
		}

		// проверяем, нужно ли блокировать запрос
		isBlock, err := check_request.IsBlock(r, webAppService, policyService, ruleService, actionService, actionExecutor)
		if err != nil {
			fmt.Printf("failed to check request %s", err)
			return
		}
		if isBlock {
			fmt.Println("access will be denied")
			http.Error(w, "Access Denied", http.StatusForbidden)
			return
		} else {
			fmt.Println("access wil not be denied")
		}

		r.Header.Set("X-Proxy-Port", fmt.Sprintf("%d", port))

		webApp, err := webAppService.GetWebAppForHost(r.Context(), host)
		if err != nil {
			fmt.Printf("failed to get web app for host %s, %s", r.Host, err)
			// может тут надо блок? или ошибку 5**
			return
		}
		proxy := wafConfig.GetProxyForWebApp(webApp)
		fmt.Printf("Forward request %s %s to upstream", r.Method, r.URL.Path)
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

	closeAll(blackList, errorLogConfig, accessLogConfig)

	log.Println("Server stopped")
}

// закрываем нужные ресурсы - файл для лога и гео базу
func closeAll(bl bl.Blacklist, errorLogConfig, accessLogConfig *log_config.LogConfig) {
	errorLogConfig.CloseLogFile()
	accessLogConfig.CloseLogFile()
	geo.CloseGeoDB() // подумать, может как-то по-лучше организовать работу с гео
	bl.Close()
}
