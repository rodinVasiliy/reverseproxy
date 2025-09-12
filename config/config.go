package config

// key - domain value - webapplication
type Config struct {
	Configs map[string]*WebApp
}

type SSLConfiguration struct {
	certPath string
	keyPath  string
}

type Rule struct {
	name       string
	parameters []string
	actions    []Action
}

type WebApp struct {
	pol              *Policy
	upstream         string
	sslConfiguration *SSLConfiguration
}

func (c *Config) Add(domain string, webapplication *WebApp) {
	// TODO добавить проверку на пустую мапу
	c.Configs[domain] = webapplication
}

func (c *Config) Get(domain string) *WebApp {
	val, ok := c.Configs[domain]
	if ok {
		return val
	} else {
		return nil
	}
}
