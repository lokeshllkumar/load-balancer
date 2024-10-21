package utils

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Backends      []BackendConfig `yaml:"backends"`
	Health        HealthConfig    `yaml:"health"`
	LoadBalancing LBConfig        `yaml:"load_balancing"`
}

type BackendConfig struct {
	URL    string `yaml:"url"`
	Weight int    `yaml:"weight"`
	Sticky bool   `yaml:"sticky"`
}

type HealthConfig struct {
	Interval int    `yaml:"interval"`
	Timeout  int    `yaml:"timeout"`
	Retries  int    `yaml:"retries"`
	Path     string `yaml:"path"`
}

type LBConfig struct {
	Strategy string `yaml:"strategy"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
