package balancer

import (
	"context"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/lokeshllkumar/load-balancer/internal/healthcheck"
	"github.com/lokeshllkumar/load-balancer/internal/metrics"
	"github.com/lokeshllkumar/load-balancer/internal/registry"
)

type Backend struct {
	URL         *url.URL
	Alive       bool
	Connections int32
	mux         sync.RWMutex // keeping it private
	LastError   time.Time
	ErrorCount  int
	HealthPath  string
	InstanceID  string
}

func (b *Backend) SetAlive(alive bool) {
	b.mux.Lock()
	b.Alive = alive
	if alive {
		b.ErrorCount = 0
	}
	b.mux.Unlock()
}

func (b *Backend) IsAlive() bool {
	b.mux.RLock()
	alive := b.Alive
	b.mux.RUnlock()
	return alive
}

func (b *Backend) IncrementConnections() {
	b.mux.Lock()
	b.Connections++
	metrics.ActiveConnectionsGauge.WithLabelValues(b.URL.Host, b.InstanceID).Set(float64(b.Connections))
	b.mux.Unlock()
}

func (b *Backend) DecrementConnections() {
	b.mux.Lock()
	if b.Connections > 0 {
		b.Connections--
	}
	metrics.ActiveConnectionsGauge.WithLabelValues(b.URL.Host, b.InstanceID).Set(float64(b.Connections))
	b.mux.Unlock()
}

func (b *Backend) GetConnections() int32 {
	b.mux.RLock()
	conn := b.Connections
	b.mux.RUnlock()
	return conn
}

func (b *Backend) RecordError() {
	b.mux.Lock()
	b.ErrorCount++
	b.LastError = time.Now()
	// simple circuit breaker - if there any too many errors in a short span of time, mark unhealthy
	if b.ErrorCount > 4 && time.Since(b.LastError) < 10*time.Second {
		if b.Alive {
			b.Alive = false
			log.Printf("Backend %s (%s) marked unhealthy due to repeated errors", b.URL.String(), b.InstanceID)
			metrics.ActiveConnectionsGauge.WithLabelValues(b.URL.Host, b.InstanceID).Set(0)
		}
	}
	b.mux.Unlock()
}

type BackendManager struct {
	backends           []*Backend
	mu                 sync.RWMutex
	serviceRegistry    registry.ServiceRegistryClient
	healthCheckTicker  *time.Ticker
	discoveryTicker    *time.Ticker
	healthCheckTimeout time.Duration
	stopChan           chan struct{}
}

func NewBackendManager(serviceRegistryClient registry.ServiceRegistryClient, healthCheckInterval string, healthCheckTimeout string) *BackendManager { // store the interval and timeout as strings
	hInterval, err := time.ParseDuration(healthCheckInterval)	
	if err != nil {
		log.Fatalf("Invalid health check interval duration: %v", err)
	}
	hTimeout, err := time.ParseDuration(healthCheckTimeout)
	if err != nil {
		log.Fatalf("Invalid health check timeout duration: %v", err)
	}

	return &BackendManager{
		backends: make([]*Backend, 0),
		serviceRegistry: serviceRegistryClient,
		healthCheckTicker: time.NewTicker(hInterval),
		discoveryTicker: time.NewTicker(hInterval * 2),
		healthCheckTimeout: hTimeout,
		stopChan: make(chan struct{}),
	}
}

func (bm *BackendManager) StartBackendDiscovery(ctx context.Context) {
	log.Println("Starting backedn discovery...")
	bm.discoverBackends()
	for {
		select {
		case <- bm.discoveryTicker.C:
			bm.discoverBackends()
		case <- bm.stopChan:
			log.Println("Backend discovery stopped")
			return
		case <- ctx.Done():
			log.Println("Backend discovery cotnext cancelled")
			return
		}
	}
}

func (bm *BackendManager) discoverBackends() {
	log.Println("Discovering backends from service registry...")
	registeredServices, err := bm.serviceRegistry.GetServices()
	if err != nil {
		log.Printf("Failed to fetch services frpm registry: %v", err)
		return
	}

	newBackends := make([]*Backend, 0, len(registeredServices))
	existingBackendsMap := make(map[string]*Backend)

	bm.mu.RLock()
	for _, b := range bm.backends {
		existingBackendsMap[b.InstanceID] = b
	}
	bm.mu.RUnlock()

	for _, s := range registeredServices {
		backendURL, err := url.Parse(s.URL)
		if err != nil {
			log.Printf("Invalid backend URL received from registry: %s, error: %v", s.URL, err)
			continue
		}

		if existingBackend, found := existingBackendsMap[s.ID]; found {
			// yet to implement backend updation if props change
			newBackends = append(newBackends, existingBackend)
			delete(existingBackendsMap, s.ID) // cleanup
		} else {
			newBackend := &Backend{
				URL: backendURL,
				Alive: false,
				HealthPath: s.HealthPath,
				InstanceID: s.ID,
			}
			newBackends = append(newBackends, newBackend)
			log.Printf("Discovered new backend: %s (ID: %s)", newBackend.URL.String(), newBackend.InstanceID)
		}
	}

	bm.mu.Lock()
	bm.backends = newBackends
	bm.mu.Unlock()

	// cleaning up deregsitered/unresponsive backends
	for _, removedBackend := range existingBackendsMap {
		log.Printf("Backend %s (ID: %s) removed (deregistered or no longer reported)", removedBackend.URL.String(), removedBackend.InstanceID)
		metrics.BackendStatusGauge.WithLabelValues(removedBackend.URL.Host, removedBackend.InstanceID).Set(0)
		metrics.ActiveConnectionsGauge.WithLabelValues(removedBackend.URL.Host, removedBackend.InstanceID).Set(0)
	}
	log.Printf("Finished backend discovery. There are currently %d backends available", len(bm.backends))
}

func(bm *BackendManager) StartHealthChecks(ctx context.Context) {
	log.Println("Starting backend health checks...")
	for {
		select {
		case <- bm.healthCheckTicker.C:
			bm.checkAllBackends()
		case <- bm.stopChan:
			log.Println("Health checks stopped")
			return
		case <- ctx.Done():
			log.Println("Health checks context cancelled")
			return
		}
	}
}

func (bm *BackendManager) checkAllBackends() {
	bm.mu.RLock()
	backendsToCheck := make([]*Backend, len(bm.backends))
	copy(backendsToCheck, bm.backends)
	bm.mu.RUnlock()

	for _, backend := range backendsToCheck {
		go bm.performHealthCheck(backend)
	}
}

func (bm *BackendManager) performHealthCheck(backend *Backend) {
	fullHealthURL := backend.URL.String() + backend.HealthPath
	isHealthy := healthcheck.CheckHTTP(fullHealthURL, bm.healthCheckTimeout)

	if isHealthy {
		if !backend.IsAlive() {
			log.Printf("Backend %s (ID: %s) is now healthy", backend.URL.String(), backend.InstanceID)
			backend.SetAlive(true)
			metrics.BackendStatusGauge.WithLabelValues(backend.URL.Host, backend.InstanceID).Set(1)
		}
	} else {
		if backend.IsAlive() {
			log.Printf("Backend %s (ID: %s) is now unhealthy", backend.URL.String(), backend.InstanceID)
			backend.SetAlive(false)
			metrics.BackendStatusGauge.WithLabelValues(backend.URL.Host, backend.InstanceID).Set(0)
		}
	}
}

func (bm *BackendManager) GetHealthyBackends() []*Backend {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	healthyBackends := make([]*Backend, 0)
	for _, b := range bm.backends {
		if b.IsAlive() {
			healthyBackends = append(healthyBackends, b)
		}
	}

	return healthyBackends
}

func (bm *BackendManager) Stop() {
	bm.healthCheckTicker.Stop()
	bm.discoveryTicker.Stop()
	close(bm.stopChan)
}