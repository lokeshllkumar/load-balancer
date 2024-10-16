package balancer

import (
	"sync"

	"github.com/lokeshllkumar/load-balancer/internal/backend"
)

type StickySessions struct {
	sessions map[string]*backend.Backend
	mu       sync.RWMutex
}

func NewStickySession() *StickySessions {
	return &StickySessions{sessions: make(map[string]*backend.Backend)}
}

func (s *StickySessions) GetBackendSS(sessionID string, backends []*backend.Backend) *backend.Backend {
	s.mu.RLock()
	if backend, exists := s.sessions[sessionID]; exists {
		s.mu.RUnlock()
		return backend
	}
	s.mu.RUnlock()

	backend := NewRoundRobin().GetNextBackendRR(backends)
	s.mu.Lock()
	s.sessions[sessionID] = backend
	s.mu.Unlock()

	return backend
}
