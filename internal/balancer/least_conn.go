package balancer

import (
	"github.com/lokeshllkumar/load-balancer/internal/backend"
)

type LeastConnections struct{}

func NewLeastConn() *LeastConnections {
	return &LeastConnections{}
}

func (lc *LeastConnections) GetNextBackendLC(backends []*backend.Backend) *backend.Backend {
	var least *backend.Backend
	for _, b := range backends {
		if least == nil || b.Connections < least.Connections {
			least = b
		}
	}

	return least
}
