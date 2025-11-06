package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type MongoConfig struct {
	URI      string `yaml:"uri" json:"uri"`
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	Database string `yaml:"database" json:"database"`
	User     string `yaml:"user" json:"user"`
	Password string `yaml:"password" json:"password"`
	IsMaster bool   `yaml:"isMaster" json:"isMaster"`
}

func LoadConfig(path string) (*MongoConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg struct {
		Mongo MongoConfig `yaml:"mongo"`
	}

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg.Mongo, nil
}
