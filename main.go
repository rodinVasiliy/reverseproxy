package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	appPolicyService "reverseproxy/internal/app/policy"
	appRule "reverseproxy/internal/app/rule"
	appSSLService "reverseproxy/internal/app/ssl"
	appWebapp "reverseproxy/internal/app/webapp"
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
	wafAction "reverseproxy/internal/waf/action"
	parsedRequest "reverseproxy/internal/waf/parsed_request"
	wafPolicy "reverseproxy/internal/waf/policy"
	wafRule "reverseproxy/internal/waf/rule"

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

func fail(msg string, err error) {
	fmt.Printf("%s: %v", msg, err) // позже уберем
	log.Fatalf("%s: %v", msg, err)
}

func getInItFlag() bool {
	return os.Getenv("INIT") == "1"
}

func startAdminAPI(url string, actionService *actionDoc.Service, policyService *policy.Service, sslService *ssl.Service,
	webAppService *webapp.Service, appWebappService *appWebapp.AppWebappService,
	appSSLService *appSSLService.AppSSLService, appPolicyService *appPolicyService.AppPolicyService,
	ruleService *rule.Service, appRuleService *appRule.AppRuleService, manager *manager.Manager) {
	adminRouter := gin.Default()
	adminRouter.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Authorization"},
	}))
	api := adminRouter.Group("/admin/api")
	handler.RegisterActionRoutes(api, actionService)
	handler.RegisterPolicyRoutes(api, policyService, appPolicyService)
	handler.RegisterSSLRoutes(api, sslService, appSSLService)
	handler.RegisterWebAppRoutes(api, webAppService, appWebappService, manager)
	handler.RegisterRuleRoutes(api, ruleService, appRuleService)

	go func() {
		if err := adminRouter.Run(url); err != nil {
			log.Printf("admin api stopped: %v", err)
		}
	}()
}

func main() {

	// fmt в консольку для информации - вся инфа до такого как все запустилось
	fmt.Println("reverse proxy ...")

	nodeConfig, err := node.GetNodeConfig()
	if err != nil {
		fail("failed to get node config", err)
		return
	}
	// Порт прокси(waf), куда nginx будет отправлять запросы
	port := nodeConfig.Port

	errorLogFileName := filepath.Join("log", "error.log")
	eventsLogFileName := filepath.Join("log", "events.log")
	accessLogFileName := filepath.Join("log", "access.log")

	// Error log - для записи всех ошибок
	errorLogConfig, err := log_config.NewLogConfig(errorLogFileName)
	if err != nil {
		fail("failed to open error log file", err)
		return
	}
	defer errorLogConfig.CloseLogFile()
	log.SetOutput(errorLogConfig.File())

	// Access log - для записи всех дошедших до WAF запросов
	accessLogConfig, err := log_config.NewLogConfig(accessLogFileName)
	if err != nil {
		fail("failed to open access log file", err)
		return
	}
	defer accessLogConfig.CloseLogFile()
	accessLogger := log.New(accessLogConfig.File(), "", log.LstdFlags|log.Lmicroseconds)

	eventLogConfig, err := log_config.NewLogConfig(eventsLogFileName)
	if err != nil {
		fail("failed to open event log file", err)
		return
	}
	defer eventLogConfig.CloseLogFile()
	eventsLogger := log.New(eventLogConfig.File(), "", log.LstdFlags|log.Lmicroseconds)

	var blackList *bl.RedisBL
	blackList, err = bl.NewRedisBL()
	if err != nil {
		fail("failed to init blacklist", err)
		return
	}
	defer blackList.Close()

	// Загружаем гео базу
	fmt.Println("loading geo base from file ...")
	err = geo.InitGeo()
	if err != nil {
		fail("failed to init geo base", err)
		return
	}
	defer geo.CloseGeoDB()

	fmt.Println("Getting mongo db dependencies ...")
	// Подключаемся к монгодб
	mongoDeps, err := config.NewMongoDeps()
	if err != nil {
		fail("failed to get mongo deps", err)
		return
	}

	fmt.Println("Config Initialization started...")

	actionRepository := repository.NewMongoRepository[actionDoc.ActionDoc](mongoDeps.Client, repository.DB_NAME, repository.ACTION_COLLECTION)
	actionService := actionDoc.NewService(actionRepository)

	// Сами actions, зашитые в код. Передаем туда eventsLogger - cюда будут записываться сработки по правилам.
	actionRegistry := wafAction.BuildRegistry(eventsLogger)

	ruleRepository := repository.NewMongoRepository[rule.Rule](mongoDeps.Client, repository.DB_NAME, repository.RULE_COLLECTION)
	ruleService := rule.NewService(ruleRepository)

	policyRepository := repository.NewMongoRepository[policy.Policy](mongoDeps.Client, repository.DB_NAME, repository.POLICY_COLLECTION)
	policyService := policy.NewService(policyRepository)

	sslRepository := repository.NewMongoRepository[ssl.SSL](mongoDeps.Client, repository.DB_NAME, repository.SSL_COLLECTION)
	sslService := ssl.NewService(sslRepository)

	webappRepository := repository.NewMongoRepository[webapp.WebApp](mongoDeps.Client, repository.DB_NAME, repository.WEBAPP_COLLECTION)
	webappService := webapp.NewService(webappRepository)
	webappSyncService := appWebapp.NewWebappSyncService(sslService)

	watcher := webapp.NewWatcher(webappRepository, webappSyncService)
	go watcher.Watch(context.Background())

	// Конфиг нужен, чтобы для хоста выдавать httputil.ReverseProxy
	manager, err := manager.NewManager(webappService)
	if err != nil {
		fail("failed to load waf config", err)
		return
	}
	fmt.Println("Waf Config successfully loaded")

	////////////////////// Инициализация БД //////////////////////
	if getInItFlag() {
		fmt.Println("Initialization database ...")
		utils.ClearAllCollections(mongoDeps)
		utils.DropOldWebappFiles()
		err = initialization.InItDB(policyService, actionService, ruleService)
		if err != nil {
			fail("failed to init db", err)
			return
		}
		err = initialization.NewTestWebApp(policyService, sslService, webappService, manager)
		if err != nil {
			fail("failed to add test webapp", err)
			return
		}
	} else {
		fmt.Println("Init db not required")
	}

	////////////////////// Компиляция правил, политик, действий //////////////////////
	actionDocs, err := actionService.FindAll(context.Background())
	if err != nil {
		fail("failed to find all actions", err)
		return
	}
	actionEngine := &wafAction.ActionEngine{}
	err = actionEngine.Load(actionDocs, actionRegistry)
	if err != nil {
		fail("failed to compile actions", err)
		return
	}

	rules, err := ruleService.FindAll(context.Background())
	if err != nil {
		fail("failed to find all rules", err)
		return
	}
	ruleEngine := &wafRule.RuleEngine{}
	ruleEngine.SetActionEngine(actionEngine)

	err = ruleEngine.Load(rules)
	if err != nil {
		fail("failed to compile rules", err)
		return
	}

	policies, err := policyService.FindAll(context.Background())
	if err != nil {
		fail("failed to find all policies", err)
		return
	}
	policyEngine := &wafPolicy.PolicyEngine{}
	policyEngine.SetActionEngine(actionEngine)
	policyEngine.SetRuleEngine(ruleEngine)
	err = policyEngine.Load(policies)
	if err != nil {
		fail("failed to compile policies %s", err)
		return
	}

	////////////////////// Application services //////////////////////
	appWebappService := appWebapp.NewService(webappService, policyService, sslService)
	appPolicyService := appPolicyService.NewAppPolicyService(policyService, actionService, ruleService, webappService, policyEngine)
	appSSLService := appSSLService.NewAppSSLConfiguration(sslService, webappService)
	appRuleService := appRule.NewAppRuleService(ruleService, actionService, policyService, ruleEngine, policyEngine)

	fmt.Println("starting admin api")
	startAdminAPI(nodeConfig.AdminURL, actionService, policyService, sslService, webappService, appWebappService,
		appSSLService, appPolicyService, ruleService, appRuleService, manager)
	/////////////////////////////////////////////////////////////////

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Proxy request %s %s %s via port %d\n", r.Host, r.Method, r.URL.Path, port)

		host := r.Host
		ip := utils.GetIpFromRequest(r)
		ok, err := blackList.Exists(ip.String())
		if err != nil {
			fail("failed to check request in BL", err)
		}
		if ok {
			fmt.Println("access will be denied")
			accessLogger.Printf("403\t%s\t%s\t%s\t%v\n", r.Host, r.Method, r.URL.Path, port)

			http.Error(w, "Access Denied", http.StatusForbidden)
			return
		}

		// Проверяем, нужно ли блокировать запрос
		wa, err := webappService.GetWebAppForHost(r.Context(), host)
		// Возвращаем 404, если для запроса не удалось найти конфигурацию
		if err != nil {
			accessLogger.Printf("404\t%s\t%s\t%s\t%v\n", r.Host, r.Method, r.URL.Path, port)
			http.Error(w, "Not found", http.StatusNotFound)
			fmt.Printf("failed to find webapp for host %s; host:%s\n", err, host)
			return
		}

		policyId := wa.PolicyId
		compiledPolicy, ok := policyEngine.Get(policyId)
		if !ok {
			accessLogger.Printf("404\t%s\t%s\t%s\t%v\n", r.Host, r.Method, r.URL.Path, port)
			http.Error(w, "Not found", http.StatusNotFound)
			fmt.Printf("failed to find compiled policy by id %s\n", err)
			return
		}

		parsedRequest := parsedRequest.NewParsedRequest(r)
		isBlock, err := compiledPolicy.Evaluate(parsedRequest, eventsLogger, blackList)
		if err != nil {
			accessLogger.Printf("502\t%s\t%s\t%s\t%v\n", r.Host, r.Method, r.URL.Path, port)
			http.Error(w, "Internal error", http.StatusInternalServerError)

			// TO DO сложить в fail
			log.Printf("failed to check request %s\t%s\t%s\t%v\tError:%v", r.Host, r.Method, r.URL.Path, port, err)
			fmt.Printf("failed to check request %s\t%s\t%s\t%v\tError:%v", r.Host, r.Method, r.URL.Path, port, err)
			return
		}
		if isBlock {
			fmt.Println("access will be denied")
			accessLogger.Printf("403\t%s\t%s\t%s\t%v\n", r.Host, r.Method, r.URL.Path, port)
			http.Error(w, "Access Denied", http.StatusForbidden)
			return
		} else {
			fmt.Println("access wil not be denied")
		}

		r.Header.Set("X-Proxy-Port", fmt.Sprintf("%d", port))

		proxy, ok := manager.GetProxyForWebApp(wa)
		if !ok {
			accessLogger.Printf("502\t%s\t%s\t%s\t%v\n", r.Host, r.Method, r.URL.Path, port)
			http.Error(w, "Internal error", http.StatusInternalServerError)

			// TO DO сложить в fail
			log.Printf("Failed to get proxy for request %s\t%s\t%s\t%v\tError:%v", r.Host, r.Method, r.URL.Path, port, err)
			fmt.Printf("Failed to get proxy for request %s\t%s\t%s\t%v\tError:%v", r.Host, r.Method, r.URL.Path, port, err)
			return
		}
		accessLogger.Printf("forward request to upstream: %s\t%s\t%s\t%v\n", r.Host, r.Method, r.URL.Path, port)
		fmt.Printf("Forward request %s %s to upstream\n", r.Method, r.URL.Path)
		proxy.ServeHTTP(w, r)
	})

	// Слушаем только с nginx, port - порт waf, на который ему nginx пересылает запросы(nginx слушает 443 порт, а на WAF отправляет на 4443, например)
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
			fail("failed to run server", err)
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

	log.Println("Server stopped")
}
