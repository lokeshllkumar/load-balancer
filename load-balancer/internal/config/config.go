package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Port                int    `yaml:"port"`
	Strategy            string `yaml:"strategy"`
	ServiceRegsistryUrl string `yaml:"serviceRegistryURL"`
	ServiceRegistryType string `yaml:"serviceRegistryType"`
	HealthCheckInterval string `yaml:"healthCheckInterval"`
	BackendHealthPath   string `yaml:"backendHealthPath"`
	HealthCheckTimeout  string `yaml:"healthCheckTimeout"` // can change to float32
}

func LoadConfig(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config object: %v", err)
	}

	return &cfg, nil
}
