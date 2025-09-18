package config

import (
	"fmt"
	"log"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
)

// key - domain value - webapplication
type Config struct {
	configs map[string]*WebApp
	logFile *os.File
}

type SSLConfiguration struct {
	certPath string
	keyPath  string
}

func (c *Config) Add(domain string, webapplication *WebApp) {
	// TODO добавить проверку на пустую мапу
	c.configs[domain] = webapplication
}

func (c *Config) Get(domain string) *WebApp {
	val, ok := c.configs[domain]
	if ok {
		return val
	} else {
		return nil
	}
}

// в дальнейшем сделать добавление из веба, а потом добавить импорт/экспорт настроек
// на данном этапе чисто тестовый вариант
func InitConfig(port int) (*Config, error) {

	logFile := initLogFile(port)

	cfg := make(map[string]*WebApp, 5)
	domain := "myproxytest.site"
	upstream, err := url.Parse("http://localhost:9091")
	if err != nil {
		return nil, fmt.Errorf("failed to parse upstream")
	}
	policy := DefaultPolicy()
	// мб экранировать
	ssl := SSLConfiguration{
		certPath: "/etc/letsencrypt/live/myproxytest.site/fullchain.pem",
		keyPath:  "/etc/letsencrypt/live/myproxytest.site/privkey.pem",
	}
	var webApp = WebApp{&policy, upstream, &ssl, httputil.NewSingleHostReverseProxy(upstream)}
	cfg[domain] = &webApp
	return &Config{configs: cfg, logFile: logFile}, nil
}

func GetUpstreams(cfg *Config) []*url.URL {
	var result []*url.URL
	for _, webApp := range cfg.configs {
		if webApp == nil || webApp.upstream == nil {
			continue
		}
		result = append(result, webApp.upstream)
	}
	return result
}

func (cfg *Config) GetReverseProxyForHost(domain string) *httputil.ReverseProxy {
	wa := cfg.configs[domain]
	return wa.proxy
}

func (cfg *Config) GetPolicyForHost(domain string) *Policy {
	wa := cfg.configs[domain]
	return wa.pol
}

func initLogFile(port int) *os.File {
	logFileName := filepath.Join("log", fmt.Sprintf("db_%d.log", port))
	var err error
	var logFile *os.File
	logFile, err = os.OpenFile(logFileName,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Failed to open log file: %v", err)
	}

	// Настраиваем логгер на запись в файл
	log.SetOutput(logFile)
	return logFile
}

func (cfg *Config) CloseLogFile() {
	if cfg.logFile != nil {
		cfg.logFile.Close()
		fmt.Println("log file closed")
	}
}
