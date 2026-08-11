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
	wafPolicy "reverseproxy/internal/waf/policy"
	manager "reverseproxy/internal/waf/proxy"
	wafRule "reverseproxy/internal/waf/rule"
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
	appRuleSerive    *appRule.AppRuleService
}

func InItServices() (*Services, error) {

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

	// TO DO - подумать, чтобы добавить в cancel
	watcher := webapp.NewWatcher(webappRepository, webappSyncService)
	go watcher.Watch(context.Background())

	// Конфиг нужен, чтобы для хоста выдавать httputil.ReverseProxy
	manager, err := manager.NewManager(webappService)
	if err != nil {
		return nil, fmt.Errorf("failed get new manager %w", err)
	}
	fmt.Println("Waf Config successfully loaded")

	////////////////////// Компиляция правил, политик, действий //////////////////////
	actionDocs, err := actionService.FindAll(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get all actions %w", err)
	}
	actionEngine := &wafAction.ActionEngine{}
	err = actionEngine.Load(actionDocs, actionRegistry)
	if err != nil {
		return nil, fmt.Errorf("failed to compile actions %w", err)
	}

	rules, err := ruleService.FindAll(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get all rules %w", err)
	}
	ruleEngine := &wafRule.RuleEngine{}
	ruleEngine.SetActionEngine(actionEngine)

	err = ruleEngine.Load(rules)
	if err != nil {
		return nil, fmt.Errorf("failed to compile rules %w", err)
	}

	policies, err := policyService.FindAll(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get all policies %w", err)
	}
	policyEngine := &wafPolicy.PolicyEngine{}
	policyEngine.SetActionEngine(actionEngine)
	policyEngine.SetRuleEngine(ruleEngine)
	err = policyEngine.Load(policies)
	if err != nil {
		return nil, fmt.Errorf("failed to compile policies %w", err)
	}

	////////////////////// Application services //////////////////////

	appWebappService := appWebapp.NewService(webappService, policyService, sslService)
	appPolicyService := appPolicyService.NewAppPolicyService(policyService, actionService, ruleService, webappService, policyEngine)
	appSSLService := appSSLService.NewAppSSLConfiguration(sslService, webappService)
	appRuleService := appRule.NewAppRuleService(ruleService, actionService, policyService, ruleEngine, policyEngine)

	/////////////////////////////////////////////////////////////////
	services := &Services{
		mongoDeps:         mongoDeps,
		accessLogger:      accessLogger,
		errorLogger:       errorLogger,
		eventLogger:       eventsLogger,
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
		appWebappService:  appWebappService,
		appPolicyService:  appPolicyService,
		appSSLService:     appSSLService,
		appRuleSerive:     appRuleService,
	}
	return services, nil
}

func (service *Services) Close() {
	if service.blackList != nil {
		service.blackList.Close()
	}

	geo.CloseGeoDB()

	service.accessLogConfig.CloseLogFile()
	service.errorLogConfig.CloseLogFile()
	service.eventLogConfig.CloseLogFile()
}
