package config

import (
	"fmt"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
)

// key - domain value - webapplication
type Config struct {
	configs map[string]*WebApp // мапа - ключ - домен, значение - веб приложение
	logFile *os.File           // файл для лога, у каждой прокси он будет свой, путь зависит от порта на которой работает прокси
	blFile  *BL
}

type SSLConfiguration struct {
	certPath string
	keyPath  string
}

// в дальнейшем сделать добавление из веба, а потом добавить импорт/экспорт настроек
// на данном этапе чисто тестовый вариант
func InitConfig(port int) (*Config, error) {
	logFile, err := initLogFile(port)
	if err != nil {
		return nil, err
	}

	cfg := make(map[string]*WebApp, 5) // пока ограничимся 5-ю приложениями
	domain := "myproxytest.site"

	cfg[domain], err = initWebApp()
	if err != nil {
		return nil, err
	}

	fmt.Println("Loading BL file")
	blPath := filepath.Join("config", "blacklist.db")
	bl, err := NewBlacklistStore(blPath)
	bl.Add("") // TODO добавить ip + протестировать BL
	if err != nil {
		return nil, err
	}
	fmt.Println("BL file successfully loaded")

	return &Config{configs: cfg, logFile: logFile, blFile: bl}, nil
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
