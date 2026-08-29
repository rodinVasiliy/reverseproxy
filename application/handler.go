package application

import (
	"fmt"
	"net/http"
	"reverseproxy/internal/domain/webapp"
	"reverseproxy/internal/utils"
	"reverseproxy/internal/waf/parsedrequest"
	"reverseproxy/internal/waf/policy"
)

func getHandler(application *Application) http.HandlerFunc {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestInfo := fmt.Sprintf("request %s %s %s via port %d\n", r.Host, r.Method, r.URL.Path, application.nodeConfig.Port)
		fmt.Print(requestInfo)
		host := r.Host
		ip := utils.GetIpFromRequest(r)

		// CHECK IN BL
		ok, err := checkInBL(application, ip.String())
		if err != nil {
			// Пока ошибку игнорируем, не хочется положить всё, если BL упадет или что-то в этом роде.
			application.services.errorLogger.Printf("failed to check ip %s in BL: %s", ip, err)
		} else {
			if ok {
				application.services.eventLogger.Printf("Blacklist\t403\t%s", requestInfo)
				application.services.accessLogger.Printf("403\t%s", requestInfo)
				deny(w)
				return
			}
		}

		// GET WA
		webapp, err := getWebapp(application, r, host)
		if !ok {
			application.services.accessLogger.Printf("404\t%s", requestInfo)
			application.services.errorLogger.Printf("failed to find webapp for host: %s; host:%s\n", err, host)
			notFound(w) // Возвращаем 404, если для запроса не удалось найти webapp
			return
		}

		// GET POLICY
		compiledPolicy, ok := getPolicy(application, webapp)
		if !ok {
			application.services.accessLogger.Printf("404\t%s", requestInfo)
			application.services.errorLogger.Printf("failed to find compiled policy by id: %s\n", compiledPolicy.ID.String())
			notFound(w) // Возвращаем 404, если для запроса не удалось найти policy для webapp
			return
		}

		// CHECK TO BLOCK
		isBlock, err := checkBlock(application, r, compiledPolicy)
		if err != nil {
			application.services.accessLogger.Printf("502\t%s", requestInfo)
			application.services.errorLogger.Printf("failed to check request: %s", requestInfo)
			internalError(w)
			return
		}
		if isBlock {
			application.services.accessLogger.Printf("403\t%s", requestInfo)
			deny(w)
			return
		}

		// SEND TO UPSTREAM
		if !sendToUpstream(application, w, r, webapp, requestInfo) {
			application.services.accessLogger.Printf("502\t%s", requestInfo)
			application.services.errorLogger.Printf("failed to get proxy for webapp %s", webapp.ID)
			internalError(w)
			return
		}
	})
	return handler
}

func checkInBL(application *Application, ip string) (bool, error) {
	return application.services.blackList.Exists(ip)
}

func getWebapp(application *Application, request *http.Request, host string) (*webapp.WebApp, error) {
	return application.services.webappService.GetWebAppForHost(request.Context(), host)
}

func getPolicy(application *Application, webapp *webapp.WebApp) (*policy.CompiledPolicy, bool) {
	return application.compiler.PolicyCompiler.Get(webapp.PolicyId)
}

func checkBlock(application *Application, request *http.Request, compiledPolicy *policy.CompiledPolicy) (bool, error) {
	parsedRequest := parsedrequest.NewParsedRequest(request)
	return compiledPolicy.Evaluate(parsedRequest, application.services.eventLogger, application.services.blackList)
}

func sendToUpstream(application *Application, w http.ResponseWriter, request *http.Request, webapp *webapp.WebApp, requestInfo string) bool {
	request.Header.Set("X-Proxy-Port", fmt.Sprintf("%d", application.nodeConfig.Port))
	proxy, ok := application.services.manager.GetProxyForWebApp(webapp)
	if !ok {
		return false
	}
	application.services.accessLogger.Printf("forward request to upstream: %s", requestInfo)
	fmt.Printf("Forward request %s %s to upstream\n", request.Method, request.URL.Path)
	proxy.ServeHTTP(w, request)
	return true
}
