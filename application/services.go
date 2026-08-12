package application

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	appPolicyService "reverseproxy/internal/app/policy"
	appRule "reverseproxy/internal/app/rule"
	appSSLService "reverseproxy/internal/app/ssl"
	appWebapp "reverseproxy/internal/app/webapp"
	"reverseproxy/internal/domain/action"
	actionDoc "reverseproxy/internal/domain/action"
	policy "reverseproxy/internal/domain/policy"
	rule "reverseproxy/internal/domain/rule"
	ssl "reverseproxy/internal/domain/ssl"
	webapp "reverseproxy/internal/domain/webapp"
	bl "reverseproxy/internal/infrastructure/config/bl"
	geo "reverseproxy/internal/infrastructure/config/geo"
	log_config "reverseproxy/internal/infrastructure/config/log"
	mongoconfig "reverseproxy/internal/infrastructure/config/mongo_config"
	repository "reverseproxy/internal/infrastructure/mongo"
	wafAction "reverseproxy/internal/waf/action"
	"reverseproxy/internal/waf/compiler"
	manager "reverseproxy/internal/waf/proxy"
)

type Services struct {
	mongoDeps       *mongoconfig.MongoDeps
	accessLogger    *log.Logger
	accessLogConfig *log_config.LogConfig
	errorLogger     *log.Logger
	errorLogConfig  *log_config.LogConfig
	eventLogger     *log.Logger
	eventLogConfig  *log_config.LogConfig

	actionService  *action.Service
	actionRegistry *wafAction.Registry
	ruleService    *rule.Service
	policyService  *policy.Service
	webappService  *webapp.Service
	sslService     *ssl.Service

	webappRepository  *repository.MongoRepository[webapp.WebApp]
	webappSyncService *appWebapp.AppWebappSyncService

	blackList *bl.RedisBL

	manager *manager.Manager

	appWebappService *appWebapp.AppWebappService
	appPolicyService *appPolicyService.AppPolicyService
	appSSLService    *appSSLService.AppSSLService
	appRuleService   *appRule.AppRuleService

	compiler *compiler.Compiler
}

func inItServices() (*Services, error) {

	errorLogFileName := filepath.Join("log", "error.log")
	eventsLogFileName := filepath.Join("log", "events.log")
	accessLogFileName := filepath.Join("log", "access.log")

	// Error log - для записи всех ошибок
	errorLogConfig, err := log_config.NewLogConfig(errorLogFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to get error logger %w", err)
	}
	errorLogger := log.New(errorLogConfig.File(), "", log.LstdFlags|log.Lmicroseconds)

	// Access log - для записи всех дошедших до WAF запросов
	accessLogConfig, err := log_config.NewLogConfig(accessLogFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to get access logger %w", err)
	}
	accessLogger := log.New(accessLogConfig.File(), "", log.LstdFlags|log.Lmicroseconds)

	eventLogConfig, err := log_config.NewLogConfig(eventsLogFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to get event logger %w", err)
	}
	eventsLogger := log.New(eventLogConfig.File(), "", log.LstdFlags|log.Lmicroseconds)

	var blackList *bl.RedisBL
	blackList, err = bl.NewRedisBL()
	if err != nil {
		return nil, fmt.Errorf("failed to get blacklist %w", err)
	}

	// Загружаем гео базу
	fmt.Println("loading geo base from file ...")
	err = geo.InitGeo()
	if err != nil {
		return nil, fmt.Errorf("failed to load geo base %w", err)
	}

	// Подключаемся к монгодб
	mongoDeps, err := mongoconfig.NewMongoDeps()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongo %w", err)
	}

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

	// Конфиг нужен, чтобы для хоста выдавать httputil.ReverseProxy
	manager, err := manager.NewManager(webappService)
	if err != nil {
		return nil, fmt.Errorf("failed get new manager %w", err)
	}
	fmt.Println("Waf Config successfully loaded")

	/////////////////////////////////////////////////////////////////
	services := &Services{
		mongoDeps:         mongoDeps,
		accessLogger:      accessLogger,
		accessLogConfig:   accessLogConfig,
		errorLogger:       errorLogger,
		errorLogConfig:    errorLogConfig,
		eventLogger:       eventsLogger,
		eventLogConfig:    eventLogConfig,
		blackList:         blackList,
		manager:           manager,
		actionService:     actionService,
		actionRegistry:    actionRegistry,
		ruleService:       ruleService,
		policyService:     policyService,
		webappService:     webappService,
		sslService:        sslService,
		webappRepository:  webappRepository,
		webappSyncService: webappSyncService,
	}
	return services, nil
}

func (service *Services) CompileAll() error {
	actions, err := service.actionService.FindAll(context.Background())
	if err != nil {
		return fmt.Errorf("failed to compile all: %w", err)
	}

	rules, err := service.ruleService.FindAll(context.Background())
	if err != nil {
		return fmt.Errorf("failed to compile all: %w", err)
	}

	policies, err := service.policyService.FindAll(context.Background())
	if err != nil {
		return fmt.Errorf("failed to compile all: %w", err)
	}

	compiler, err := compiler.Compile(actions, service.actionRegistry, rules, policies)
	if err != nil {
		return fmt.Errorf("failed to compile all: %w", err)
	}

	service.compiler = compiler
	return nil
}

func (service *Services) CreateAppServices() {
	service.appWebappService = appWebapp.NewService(service.webappService, service.policyService, service.sslService)
	service.appPolicyService = appPolicyService.NewAppPolicyService(service.policyService, service.actionService, service.ruleService, service.webappService, service.compiler.PolicyCompiler)
	service.appSSLService = appSSLService.NewAppSSLConfiguration(service.sslService, service.webappService)
	service.appRuleService = appRule.NewAppRuleService(service.ruleService, service.actionService, service.policyService, service.compiler.RuleCompiler, service.compiler.PolicyCompiler)
}

func NewService() (*Services, error) {
	services, err := inItServices()
	if err != nil {
		return nil, fmt.Errorf("failed to create service: %w", err)
	}
	return services, nil
}

func (service *Services) Close() {
	if service.blackList != nil {
		service.blackList.Close()
	}

	geo.CloseGeoDB()

	if service.accessLogConfig != nil {
		service.accessLogConfig.CloseLogFile()
	}

	if service.errorLogConfig != nil {
		service.errorLogConfig.CloseLogFile()
	}

	if service.eventLogConfig != nil {
		service.eventLogConfig.CloseLogFile()
	}
}
