package backend

import (
	"errors"
	"sync"
)

type Backend struct {
	URL         string
	Alive       bool
	Connections int
}

type BackendPool struct {
	backends []*Backend
	mu       sync.RWMutex
}

func NewBackendPool() *BackendPool {
	return &BackendPool{}
}

func (p *BackendPool) AddBackend(url string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.backends = append(p.backends, &Backend{URL: url, Alive: true})
}

func (p *BackendPool) RemoveBackend(url string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	for ind, b := range p.backends {
		if b.URL == url {
			p.backends = append(p.backends[:ind], p.backends[:ind + 1]...)
			return nil
		}
	}
	return errors.New("backend not found")
}

func (p *BackendPool) GetBackends() []*Backend {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.backends
}

func (p *BackendPool) GetAliveBackends() []*Backend {
	p.mu.Lock()
	defer p.mu.Unlock()
	var alive []*Backend
	for _, b := range p.backends {
		if b.Alive {
			alive = append(alive, b)
		}
	}
	return alive
}

