package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Redis RedisConfig `yaml:"redis"`
}

type RedisConfig struct {
	MasterName    string   `yaml:"masterName"`
	Password      string   `yaml:"password"`
	DB            int      `yaml:"db"`
	SentinelAddrs []string `yaml:"sentinelAddrs"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
