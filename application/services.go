package application

import (
	"context"
	"fmt"
	"log"
	"os"
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
	mongoDeps    *mongoconfig.MongoDeps
	accessLogger *log.Logger
	errorLogger  *log.Logger
	eventLogger  *log.Logger

	actionService  *action.Service
	actionRegistry *wafAction.Registry
	ruleService    *rule.Service
	policyService  *policy.Service
	webappService  *webapp.Service
	sslService     *ssl.Service

	blackList *bl.RedisBL

	manager *manager.Manager

	appWebappService *appWebapp.AppWebappService
	appPolicyService *appPolicyService.AppPolicyService
	appSSLService    *appSSLService.AppSSLService
	appRuleSerive    *appRule.AppRuleService
}

func getInItFlag() bool {
	return os.Getenv("INIT") == "1"
}

func InItServices() *Services {

	errorLogFileName := filepath.Join("log", "error.log")
	eventsLogFileName := filepath.Join("log", "events.log")
	accessLogFileName := filepath.Join("log", "access.log")

	// Error log - для записи всех ошибок
	errorLogConfig, err := log_config.NewLogConfig(errorLogFileName)
	if err != nil {
		fail("failed to open error log file", err)
		return nil
	}
	defer errorLogConfig.CloseLogFile()
	errorLogger := log.New(errorLogConfig.File(), "", log.LstdFlags|log.Lmicroseconds)

	// Access log - для записи всех дошедших до WAF запросов
	accessLogConfig, err := log_config.NewLogConfig(accessLogFileName)
	if err != nil {
		fail("failed to open access log file", err)
		return nil
	}
	defer accessLogConfig.CloseLogFile()
	accessLogger := log.New(accessLogConfig.File(), "", log.LstdFlags|log.Lmicroseconds)

	eventLogConfig, err := log_config.NewLogConfig(eventsLogFileName)
	if err != nil {
		fail("failed to open event log file", err)
		return nil
	}
	defer eventLogConfig.CloseLogFile()
	eventsLogger := log.New(eventLogConfig.File(), "", log.LstdFlags|log.Lmicroseconds)

	var blackList *bl.RedisBL
	blackList, err = bl.NewRedisBL()
	if err != nil {
		fail("failed to init blacklist", err)
		return nil
	}
	defer blackList.Close()

	// Загружаем гео базу
	fmt.Println("loading geo base from file ...")
	err = geo.InitGeo()
	if err != nil {
		fail("failed to init geo base", err)
		return nil
	}
	defer geo.CloseGeoDB()

	// Подключаемся к монгодб
	mongoDeps, err := mongoconfig.NewMongoDeps()
	if err != nil {
		fail("failed to get mongo deps", err)
		return nil
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

	watcher := webapp.NewWatcher(webappRepository, webappSyncService)
	go watcher.Watch(context.Background())

	// Конфиг нужен, чтобы для хоста выдавать httputil.ReverseProxy
	manager, err := manager.NewManager(webappService)
	if err != nil {
		fail("failed to load waf config", err)
		return nil
	}
	fmt.Println("Waf Config successfully loaded")

	////////////////////// Компиляция правил, политик, действий //////////////////////
	actionDocs, err := actionService.FindAll(context.Background())
	if err != nil {
		fail("failed to find all actions", err)
		return nil
	}
	actionEngine := &wafAction.ActionEngine{}
	err = actionEngine.Load(actionDocs, actionRegistry)
	if err != nil {
		fail("failed to compile actions", err)
		return nil
	}

	rules, err := ruleService.FindAll(context.Background())
	if err != nil {
		fail("failed to find all rules", err)
		return nil
	}
	ruleEngine := &wafRule.RuleEngine{}
	ruleEngine.SetActionEngine(actionEngine)

	err = ruleEngine.Load(rules)
	if err != nil {
		fail("failed to compile rules", err)
		return nil
	}

	policies, err := policyService.FindAll(context.Background())
	if err != nil {
		fail("failed to find all policies", err)
		return nil
	}
	policyEngine := &wafPolicy.PolicyEngine{}
	policyEngine.SetActionEngine(actionEngine)
	policyEngine.SetRuleEngine(ruleEngine)
	err = policyEngine.Load(policies)
	if err != nil {
		fail("failed to compile policies %s", err)
		return nil
	}

	////////////////////// Application services //////////////////////

	appWebappService := appWebapp.NewService(webappService, policyService, sslService)
	appPolicyService := appPolicyService.NewAppPolicyService(policyService, actionService, ruleService, webappService, policyEngine)
	appSSLService := appSSLService.NewAppSSLConfiguration(sslService, webappService)
	appRuleService := appRule.NewAppRuleService(ruleService, actionService, policyService, ruleEngine, policyEngine)

	/////////////////////////////////////////////////////////////////
	services := &Services{
		mongoDeps:        mongoDeps,
		accessLogger:     accessLogger,
		errorLogger:      errorLogger,
		eventLogger:      errorLogger,
		blackList:        blackList,
		manager:          manager,
		actionService:    actionService,
		actionRegistry:   actionRegistry,
		ruleService:      ruleService,
		policyService:    policyService,
		webappService:    webappService,
		sslService:       sslService,
		appWebappService: appWebappService,
		appPolicyService: appPolicyService,
		appSSLService:    appSSLService,
		appRuleSerive:    appRuleService,
	}
	return services
}
