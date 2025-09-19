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

// в дальнейшем сделать добавление из веба, а потом добавить импорт/экспорт настроек
// на данном этапе чисто тестовый вариант
func InitConfig(port int) (*Config, error) {
	logFile := initLogFile(port)
	cfg := make(map[string]*WebApp, 5) // пока ограничимся 5-ю приложениями
	domain := "myproxytest.site"
	var err error
	cfg[domain], err = initWebApp()
	if err != nil {
		return nil, fmt.Errorf("init webapp error: %s", err)
	}
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

// TODO подумать, почему я вызываю это в main и делаю метод открытым...
func (cfg *Config) CloseLogFile() {
	if cfg.logFile != nil {
		cfg.logFile.Close()
		fmt.Println("log file closed")
	}
}
