package balancer

import (
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type LoadBalancingStrategy interface {
	SelectBackend(*http.Request) *Backend
	AddBackend(backend *Backend)
	RemoveBackend(backend *Backend)
}

type BackendProvider interface {
	GetHealthyBackends() []*Backend
}

// Round Robin
type StrategyRoundRobin struct {
	backends []*Backend
	current uint64
	provider BackendProvider
}

func NewRoundRobinStrategy(provider BackendProvider) *StrategyRoundRobin {
	return &StrategyRoundRobin{
		provider: provider,
	}
}

func (rr *StrategyRoundRobin) SelectBackend(req *http.Request) *Backend {
	backends := rr.provider.GetHealthyBackends()
	if len(backends) == 0 {
		return nil
	}

	idx := atomic.AddUint64(&rr.current, 1) - 1
	selected := backends[idx % uint64(len(backends))]
	log.Printf("Round Robin selected backend: %s", selected.URL.String())
	return selected
}

// no-op
func (rr *StrategyRoundRobin) AddBackend(backend *Backend) {}

func (rr *StrategyRoundRobin) RemoveBackend(backend *Backend) {}

// Least Connections
type StrategyLeastConnections struct {
	provider BackendProvider
}

func NewLeastConnectionsStrategy(provider BackendProvider) *StrategyLeastConnections {
	return &StrategyLeastConnections{
		provider: provider,
	}
}

func (lc *StrategyLeastConnections) SelectBackend(req *http.Request) *Backend {
	backends := lc.provider.GetHealthyBackends()
	if len(backends) == 0 {
		return nil
	}

	var bestBackend *Backend
	minConnections := int32(^uint32(0) >> 1)

	for _, backend := range backends {
		currentConnections := backend.GetConnections()
		if currentConnections < minConnections {
			minConnections = currentConnections
			bestBackend = backend
		}
	}
	if bestBackend != nil {
		log.Printf("Least Connections selected backend: %s (connections: %d)", bestBackend.URL.String(), bestBackend.GetConnections())
	}
	return bestBackend
}

func (lc *StrategyLeastConnections) AddBackend(backend *Backend) {}

func (lc *StrategyLeastConnections) RemoveBackend(backend *Backend) {}

// Sticky Sessions

type StrategyStickySessions struct {
	sessionMap map[string]*Backend
	mu sync.RWMutex
	provider BackendProvider
}

func NewStickySessionsStrategy(provider BackendProvider) *StrategyStickySessions {
	return &StrategyStickySessions{
		sessionMap: make(map[string]*Backend),
		provider: provider,
	}
}

func (ss *StrategyStickySessions) SelectBackend(req *http.Request) *Backend {
	sessionIDCookie, err := req.Cookie("SESSIONID")
	var sessionID string
	if err != nil {
		sessionID = sessionIDCookie.Value
	}

	if sessionID == "" {
		ss.mu.Lock()
		backend, found := ss.sessionMap[sessionID]
		ss.mu.RUnlock()

		if found && backend.IsAlive() {
			log.Printf("Sticky Session: Reusing backend %s for session %s", backend.URL.String(), sessionID)
			return backend
		} else if found && !backend.IsAlive() {
			log.Printf("Sticky Session: Backend %s for session %s is unhealthy, re-selecting", backend.URL.String(), sessionID)
			ss.mu.Lock()
			delete(ss.sessionMap, sessionID)
			ss.mu.Unlock()
		}
	}

	// init selection
	log.Println("Sticky Session: No existing session or backend unhealthy, performing initial selection")
	newBackend := NewRoundRobinStrategy(ss.provider).SelectBackend(req)

	if newBackend != nil && sessionID != "" {
		ss.mu.Lock()
		ss.sessionMap[sessionID] = newBackend
		ss.mu.Unlock()
		log.Printf("Sticky Session: New backend %s assigned to session %s", newBackend.URL.String(), sessionID)
	} else if sessionID == "" {
		newSessionID := GenerateSessionID()
		http.SetCookie(req.Context().Value("responseWriter").(http.ResponseWriter), &http.Cookie{
			Name: "SESSIONID",
			Value: newSessionID,
			Path: "/",
			Expires: time.Now().Add(24 * time.Hour),
		})
		ss.mu.Lock()
		ss.sessionMap[newSessionID] = newBackend
		ss.mu.Unlock()
		log.Printf("Sticky Session: New session ID %s created and assigned to backend %s", newSessionID, newBackend.URL.String())
	}
	return newBackend
}

func (ss *StrategyStickySessions) AddBackend(backend *Backend) {}

func (ss *StrategyStickySessions) RemoveBackend(backend *Backend) {
	ss.mu.Lock()
	for sessionID, b := range ss.sessionMap {
		if b == backend {
			delete(ss.sessionMap, sessionID)
			log.Printf("Sticky Session: Removed session %s mapping for unhealthy backend %s", sessionID, backend.URL.String())
		}
	}
	ss.mu.Unlock()
}

func GenerateSessionID() string {
	return "sess-" + time.Now().Format("20060102150405.000000")
}