package proxy

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/http/httputil"

	"github.com/lokeshllkumar/load-balancer/internal/balancer"
)

type ReverseProxyHandler struct {
	strategy balancer.LoadBalancingStrategy
}

func NewReverseProxyHandler(strategy balancer.LoadBalancingStrategy) *ReverseProxyHandler {
	return &ReverseProxyHandler{
		strategy: strategy,
	}
}

func (h *ReverseProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	strategyName := "unknown"
	switch h.strategy.(type) {
	case *balancer.StrategyRoundRobin:
		strategyName = "round_robin"
	case *balancer.StrategyLeastConnections:
		strategyName = "least_connections"
	case *balancer.StrategyStickySessions:
		strategyName = "sticky_sessions"
	}
	ctx := context.WithValue(r.Context(), "strategy_used", strategyName)
	r = r.WithContext(ctx)


	backend := h.strategy.SelectBackend(r)

	if backend == nil {
		log.Println("No healthy backend available")
		http.Error(w, "No healthy backend available", http.StatusServiceUnavailable)
		return
	}

	log.Printf("Routing request for %s to backend: %s (ID: %s) using strategy: %s", r.URL.Host, backend.URL.String(), backend.InstanceID, strategyName)

	backend.IncrementConnections()

	proxy := httputil.NewSingleHostReverseProxy(backend.URL)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
			if req.Header.Get("X-Forwarded-For") == "" {
				req.Header.Set("X-Forwarded-For", clientIP)
			}
		}
	}

	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		log.Printf("Proxy error for request %s to %s (ID: %s): %v", req.URL.Path, backend.URL.String(), backend.InstanceID, err)
		backend.RecordError()
		http.Error(rw, "Internal Server Error or Backend Unavailable", http.StatusBadGateway)
	}

	ctxWithBackendID := context.WithValue(r.Context(), "backend_id", backend.InstanceID)
	r = r.WithContext(ctxWithBackendID)

	proxy.ServeHTTP(w, r)

	backend.DecrementConnections()
}