package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	actionDoc "reverseproxy/internal/domain/action"
	policy "reverseproxy/internal/domain/policy"
	rule "reverseproxy/internal/domain/rule"
	ssl "reverseproxy/internal/domain/ssl"
	webapp "reverseproxy/internal/domain/webapp"
	bl "reverseproxy/internal/infrastructure/config/bl"
	geo "reverseproxy/internal/infrastructure/config/geo"
	log_config "reverseproxy/internal/infrastructure/config/log"
	config "reverseproxy/internal/infrastructure/config/mongo_config"
	repository "reverseproxy/internal/infrastructure/mongo"
	"reverseproxy/internal/transport/http/handler"
	action "reverseproxy/internal/waf/action"
	check_request "reverseproxy/internal/waf/check_request"

	node "reverseproxy/internal/infrastructure/config/node"
	initialization "reverseproxy/internal/initialization"
	utils "reverseproxy/internal/utils"
	manager "reverseproxy/internal/waf/proxy"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

func getInItFlag() bool {
	return os.Getenv("INIT") == "1"
}

func startAdminAPI(url string, actionService *actionDoc.Service, policyService *policy.Service, sslService *ssl.Service, webAppService *webapp.Service, manager *manager.Manager) {
	adminRouter := gin.Default()
	adminRouter.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Authorization"},
	}))
	api := adminRouter.Group("/admin/api")
	handler.RegisterActionRoutes(api, actionService)
	handler.RegisterPolicyRoutes(api, policyService)
	handler.RegisterSSLRoutes(api, sslService, webAppService)
	handler.RegisterWebAppRoutes(api, webAppService, manager)

	// to do ip брать из конфигурационного файла
	go func() {
		if err := adminRouter.Run(url); err != nil {
			log.Printf("admin api stopped: %v", err)
		}
	}()
}

// TO DO разобраться с логами
func main() {

	// fmt в консольку для информации - вся инфа до такого как все запустилось
	// log - для ошибок при работе с базой и прочим
	fmt.Println("reverse proxy ...")

	nodeConfig, err := node.GetNodeConfig()
	if err != nil {
		fmt.Printf("failed to read waf node configuration %s", err)
		return
	}
	// порт прокси(waf), куда nginx будет отправлять запросы
	port := nodeConfig.Port

	errorLogFileName := filepath.Join("log", "error.log")
	accessLogFileName := filepath.Join("log", "access.log")

	// тут идет настройка лог файла, в котором будут отображаться ошибки
	errorLogConfig, err := log_config.NewLogConfig(errorLogFileName)
	if err != nil {
		fmt.Printf("failed to open log file %s :%s\n", errorLogFileName, err)
		return
	}
	// все ошибки будут логироваться в error log
	log.SetOutput(errorLogConfig.File())
	accessLogConfig, err := log_config.NewLogConfig(accessLogFileName)
	if err != nil {
		fmt.Printf("failed to open log file %s :%s\n", accessLogFileName, err)
		closeAll(nil, errorLogConfig, nil)
		return
	}
	accessLogger := log.New(accessLogConfig.File(), "", log.LstdFlags|log.Lmicroseconds)

	var blackList *bl.RedisBL
	blackList, err = bl.NewRedisBL()
	if err != nil {
		fmt.Printf("failed to in it bl %s\n", err)
		closeAll(blackList, errorLogConfig, accessLogConfig)
		return
	}

	// загружаем гео базу
	fmt.Println("loading geo base from file ...")
	err = geo.InitGeo()
	if err != nil {
		fmt.Printf("failed to in it geo base %s\n", err)
		closeAll(blackList, errorLogConfig, accessLogConfig)
		return
	}

	fmt.Println("Getting mongo db dependencies ...")
	// подключаемся к монгодб
	mongoDeps, err := config.NewMongoDeps()
	if err != nil {
		fmt.Printf("failed to get mongo deps %s\n", err)
		closeAll(blackList, errorLogConfig, accessLogConfig)
		return
	}

	fmt.Println("Config Initialization started...")
	actionRepository := repository.NewMongoRepository[actionDoc.ActionDoc](mongoDeps.Client, repository.DB_NAME, repository.ACTION_COLLECTION)
	actionService := actionDoc.NewService(actionRepository)

	// сами actions, зашитые в код
	actionRegistry := action.NewActionRegistry(accessLogger, blackList)
	actionExecutor := action.NewExecutor(actionRegistry)

	ruleRepository := repository.NewMongoRepository[rule.Rule](mongoDeps.Client, repository.DB_NAME, repository.RULE_COLLECTION)
	ruleService := rule.NewService(ruleRepository)

	policyRepository := repository.NewMongoRepository[policy.Policy](mongoDeps.Client, repository.DB_NAME, repository.POLICY_COLLECTION)
	policyService := policy.NewService(policyRepository)

	sslRepository := repository.NewMongoRepository[ssl.SSLConfiguration](mongoDeps.Client, repository.DB_NAME, repository.SSL_COLLECTION)
	sslService := ssl.NewService(sslRepository)

	webappRepository := repository.NewMongoRepository[webapp.WebApp](mongoDeps.Client, repository.DB_NAME, repository.WEBAPP_COLLECTION)
	webAppService := webapp.NewService(webappRepository, sslService, policyService)
	go webAppService.WatchChanges() // запускаем отсмотр изменений
	// если что-то поменяется в webapp, каждая нода будет отлавливать эти изменения и создавать/удалять у себя файлы

	// конфиг нужен, чтобы для хоста выдавать httputil.ReverseProxy
	manager, err := manager.NewManager(webAppService)
	if err != nil {
		fmt.Printf("failed to load waf config %s", err)
		closeAll(blackList, errorLogConfig, accessLogConfig)
		return
	}
	fmt.Println("Waf Config successfully loaded")

	fmt.Println("starting admin api")
	startAdminAPI(nodeConfig.AdminURL, actionService, policyService, sslService, webAppService, manager)

	if getInItFlag() {
		fmt.Println("Initialization database ...")
		utils.ClearAllCollections(mongoDeps)
		utils.DropOldWebappFiles()
		go webAppService.WatchChanges()
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
		proxy := manager.GetProxyForWebApp(webApp)
		fmt.Printf("Forward request %s %s to upstream\n", r.Method, r.URL.Path)
		log.Printf("Forward request %s %s to upstream\n", r.Method, r.URL.Path)
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
