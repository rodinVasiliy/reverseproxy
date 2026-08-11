package application

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"reverseproxy/internal/domain/webapp"
	"reverseproxy/internal/infrastructure/config/node"
	"reverseproxy/internal/initialization"
	"reverseproxy/internal/transport/http/handler"
	"reverseproxy/internal/utils"
	"reverseproxy/internal/waf/compiler"
	"reverseproxy/internal/waf/parsedrequest"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Application struct {
	nodeConfig *node.NodeConfig
	services   *Services
	compiler   *compiler.Compiler

	ctx    context.Context
	cancel context.CancelFunc
}

func InitializeApplication(isNeedToInitilize bool) *Application {
	nodeConfig, err := node.GetNodeConfig()
	if err != nil {
		fail("failed to get node config", err)
		return nil
	}
	// Все сервисы
	services, err := InItServices()
	if err != nil {
		fmt.Printf("failed to init services %s", err)
		return nil
	}

	// Инициализация БД
	if isNeedToInitilize {
		initDatabase(services)
	}

	// Компилируем все сущности
	compiler, err := compileAll(services)
	if err != nil {
		fail("failed to compile", err)
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	application := &Application{
		nodeConfig: nodeConfig,
		services:   services,
		compiler:   compiler,
		ctx:        ctx,
		cancel:     cancel,
	}
	return application
}

func initDatabase(services *Services) {
	fmt.Println("Initialization database ...")
	utils.ClearAllCollections(services.mongoDeps)
	utils.DropOldWebappFiles()

	err := initialization.InItDB(services.actionService, services.ruleService, services.policyService)
	if err != nil {
		fail("failed to init db", err)
		return
	}
	err = initialization.NewTestWebApp(services.policyService, services.sslService, services.webappService, services.manager)
	if err != nil {
		fail("failed to add test webapp", err)
		return
	}
}

func (application *Application) StartAdminAPI() {
	adminRouter := gin.Default()
	adminRouter.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Authorization"},
	}))
	api := adminRouter.Group("/admin/api")
	handler.RegisterActionRoutes(api, application.services.actionService)
	handler.RegisterPolicyRoutes(api, application.services.policyService, application.services.appPolicyService)
	handler.RegisterSSLRoutes(api, application.services.sslService, application.services.appSSLService)
	handler.RegisterWebAppRoutes(api, application.services.webappService, application.services.appWebappService, application.services.manager)
	handler.RegisterRuleRoutes(api, application.services.ruleService, application.services.appRuleSerive)

	go func() {
		if err := adminRouter.Run(application.nodeConfig.AdminURL); err != nil {
			log.Printf("admin api stopped: %v", err)
		}
	}()
}

func compileAll(services *Services) (*compiler.Compiler, error) {
	ctx := context.Background()
	actions, err := services.actionService.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all actions in compile all: %w", err)
	}

	rules, err := services.ruleService.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all rules in compile all: %w", err)
	}

	policies, err := services.policyService.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all rules in compile all: %w", err)
	}

	compiler, err := compiler.Compile(actions, services.actionRegistry, rules, policies)
	if err != nil {
		return nil, fmt.Errorf("failed to compile: %w", err)
	}
	return compiler, nil
}

func (application *Application) StartProxy() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestInfo := fmt.Sprintf("request %s %s %s via port %d\n", r.Host, r.Method, r.URL.Path, application.nodeConfig.Port)
		fmt.Print(requestInfo)

		host := r.Host
		ip := utils.GetIpFromRequest(r)
		ok, err := application.services.blackList.Exists(ip.String())
		if err != nil {
			fail("failed to check request in BL", err)
		}
		if ok {
			fmt.Println("access will be denied")
			application.services.accessLogger.Printf("403\t%s", requestInfo)
			deny(w)
			return
		}

		wa, err := application.services.webappService.GetWebAppForHost(r.Context(), host)
		// Возвращаем 404, если для запроса не удалось найти конфигурацию
		if err != nil {
			application.services.accessLogger.Printf("404\t%s", requestInfo)
			notFound(w)
			fmt.Printf("failed to find webapp for host: %s; host:%s\n", err, host)
			return
		}

		policyId := wa.PolicyId
		compiledPolicy, ok := application.compiler.PolicyCompiler.Get(policyId)
		if !ok {
			application.services.accessLogger.Printf("404\t%s", requestInfo)
			notFound(w)
			fmt.Printf("failed to find compiled policy by id: %s\n", err)
			return
		}

		// Проверяем, нужно ли блокировать запрос
		parsedRequest := parsedrequest.NewParsedRequest(r)
		isBlock, err := compiledPolicy.Evaluate(parsedRequest, application.services.eventLogger, application.services.blackList)
		if err != nil {
			application.services.accessLogger.Printf("502\t%s", requestInfo)
			internalError(w)

			// TO DO сложить в fail
			msg := fmt.Sprintf("failed to check request: %s", requestInfo)
			fail(msg, err)
			return
		}
		if isBlock {
			fmt.Println("access will be denied")
			application.services.accessLogger.Printf("403\t%s", requestInfo)
			deny(w)
			return
		} else {
			fmt.Println("access wil not be denied")
		}

		r.Header.Set("X-Proxy-Port", fmt.Sprintf("%d", application.nodeConfig.Port))

		proxy, ok := application.services.manager.GetProxyForWebApp(wa)
		if !ok {
			application.services.accessLogger.Printf("502\t%s", requestInfo)
			internalError(w)
			msg := fmt.Sprintf("failed to check request: %s", requestInfo)
			fail(msg, err)
			return
		}
		application.services.accessLogger.Printf("forward request to upstream: %s", requestInfo)
		fmt.Printf("Forward request %s %s to upstream\n", r.Method, r.URL.Path)
		proxy.ServeHTTP(w, r)
	})

	// Слушаем только с nginx, port - порт waf, на который ему nginx пересылает запросы(nginx слушает 443 порт, а на WAF отправляет на 4443, например)
	addr := fmt.Sprintf("127.0.0.1:%d", application.nodeConfig.Port)

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

func (application *Application) StartWatcher() {
	watcher := webapp.NewWatcher(application.services.webappRepository, application.services.webappSyncService)
	go watcher.Watch(application.ctx)
}

func (application *Application) Close() {
	application.cancel()
	if application.services == nil {
		return
	}
	application.services.Close()
}

func (application *Application) Run() {
	application.StartAdminAPI()
	application.StartWatcher()
	application.StartProxy()
}
