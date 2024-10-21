package balancer

import (
	"errors"
	"sync"

	"github.com/lokeshllkumar/load-balancer/internal/backend"
)

type LoadBalancer struct {
	Pool      *backend.BackendPool
	clientMap sync.Map
	strategy  string
}

func NewLoadBalancer(pool *backend.BackendPool, strategy string) *LoadBalancer {
	return &LoadBalancer{
		Pool:     pool,
		strategy: strategy,
	}
}

func (lb *LoadBalancer) GetBackend(clientIP string) (*backend.Backend, error) {
	backends := lb.Pool.GetAliveBackends()
	switch lb.strategy {
	case "round-robin":
		return GetNextBackendRR(backends), nil
	case "least-connections":
		return GetNextBackendLC(backends), nil
	case "sticky":
		return lb.GetBackendSS(clientIP), nil
	default:
		return nil, errors.New("invalid load balancing strategy")
	}
}
