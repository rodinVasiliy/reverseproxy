package node

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type NodeConfig struct {
	Port     int    `yaml:"port"`
	AdminURL string `yaml:"admin_url"`
	BLURL    string `yaml:"bl_url"`
}

func GetNodeConfig() (*NodeConfig, error) {
	// TO DO подправить бы, чтобы не так убого
	path := filepath.Join("internal", "infrastructure", "config", "node", "configuration.yaml")

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg NodeConfig

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
