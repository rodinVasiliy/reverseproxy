package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type MongoConfig struct {
	URI      string `yaml:"uri" json:"uri"`
	Database string `yaml:"database" json:"database"`
	Role     string
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
