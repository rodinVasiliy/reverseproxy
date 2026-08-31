package application

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"reverseproxy/internal/domain/webapp"
	"reverseproxy/internal/infrastructure/config/node"
	"reverseproxy/internal/initialization"
	"reverseproxy/internal/transport/http/handler"
	"reverseproxy/internal/utils"
	"reverseproxy/internal/waf/compiler"
	"sync"
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

	proxyServer *http.Server
	adminServer *http.Server

	wg sync.WaitGroup
}

func InitializeApplication() *Application {
	nodeConfig, err := node.GetNodeConfig()
	if err != nil {
		fail("failed to initialize aplication: %s", err)
		return nil
	}

	services, err := NewService()
	if err != nil {
		fail("failed to get new services: %s", err)
		return nil
	}

	err = services.CompileAll()
	if err != nil {
		fail("failed to create service: %w", err)
	}

	services.CreateAppServices()

	ctx, cancel := context.WithCancel(context.Background())
	application := &Application{
		nodeConfig: nodeConfig,
		services:   services,
		compiler:   services.compiler,
		ctx:        ctx,
		cancel:     cancel,
	}

	return application
}

func initDatabase(services *Services) error {
	fmt.Println("Initialization database ...")
	utils.ClearAllCollections(services.mongoDeps)
	utils.DropOldWebappFiles()

	err := initialization.InItDB(services.actionService, services.ruleService, services.policyService)
	if err != nil {
		return fmt.Errorf("failed to initialize db: %w", err)
	}
	err = initialization.NewTestWebApp(services.policyService, services.sslService, services.webappService, services.manager)
	if err != nil {
		return fmt.Errorf("failed to add test webapp: %w", err)
	}
	return nil
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
	handler.RegisterRuleRoutes(api, application.services.ruleService, application.services.appRuleService)

	application.adminServer = &http.Server{
		Addr:    application.nodeConfig.AdminURL,
		Handler: adminRouter,
	}

	go func() {
		fmt.Printf("Starting admin API on %s\n", application.nodeConfig.AdminURL)
		err := application.adminServer.ListenAndServe()

		if err != nil && err != http.ErrServerClosed {
			application.services.errorLogger.Printf("admin api failed: %v", err)
			application.cancel()
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
	handler := getHandler(application)

	// Слушаем только с nginx, port - порт waf, на который ему nginx пересылает запросы(nginx слушает 443 порт, а на WAF отправляет на 4443, например)
	addr := fmt.Sprintf("127.0.0.1:%d", application.nodeConfig.Port)

	application.proxyServer = &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	go func() {
		fmt.Printf("Starting proxy on %s\n", addr)
		if err := application.proxyServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			application.services.errorLogger.Printf("proxy server failed: %v", err)
			application.cancel()
		}
	}()
}

func (application *Application) StartWatcher() {
	watcher := webapp.NewWatcher(
		application.services.webappRepository,
		application.services.webappSyncService,
	)

	application.wg.Add(1)

	go func() {
		defer application.wg.Done()

		watcher.Watch(application.ctx)
	}()
}

func (application *Application) Close() {
	application.cancel()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if application.proxyServer != nil {
		if err := application.proxyServer.Shutdown(ctx); err != nil {
			fmt.Printf("failed to shutdown proxy: %v\n", err) // Если context.DeadlineExceeded - пока игнорируем
		}
	}

	if application.adminServer != nil {
		if err := application.adminServer.Shutdown(ctx); err != nil {
			fmt.Printf("failed to shutdown admin api: %v\n", err) // Если context.DeadlineExceeded - пока игнорируем
		}
	}

	application.wg.Wait()

	if application.services != nil {
		application.services.Close()
	}
}

func (application *Application) Run(isNeedToInitilize bool) {
	application.StartAdminAPI()
	application.StartWatcher()
	application.StartProxy()

	if isNeedToInitilize {
		err := initDatabase(application.services)
		if err != nil {
			fail("failed to initialize aplication: %s", err)
		}
	}

	stop := make(chan os.Signal, 1)

	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(stop) // Отмена подписки на сигналы

	select {
	case sig := <-stop:
		fmt.Printf(
			"received signal %s, shutting down\n",
			sig,
		)

	case <-application.ctx.Done():
		fmt.Println("application cancelled")
	}

	application.Close()
}
