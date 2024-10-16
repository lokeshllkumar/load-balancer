package utils

import (
	"io/ioutil"

	"github.com/lokeshllkumar/load-balancer/internal/backend"
	"gopkg.in/yaml.v3"
)

func LoadBackendsFromConfig(path string) (*backend.BackendPool, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var urls []string
	if err := yaml.Unmarshal(data, &urls); err != nil {
		return nil, err
	}

	pool := backend.NewBackendPool()
	for _, url := range urls {
		pool.AddBackend(url)
	}

	return pool, nil
}