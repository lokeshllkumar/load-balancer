package balancer

import (
	"sync/atomic"

	"github.com/lokeshllkumar/load-balancer/internal/backend"
)

type RoundRobin struct {
	counter uint64
}

func NewRoundRobin() *RoundRobin {
	return &RoundRobin{}
}

func (r *RoundRobin) GetNextBackendRR(backends []*backend.Backend) *backend.Backend {
	ind := atomic.AddUint64(&r.counter, 1) % uint64(len(backends))
	return backends[ind]
}
