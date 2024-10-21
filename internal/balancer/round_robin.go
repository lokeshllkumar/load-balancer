package balancer

import (
	"github.com/lokeshllkumar/load-balancer/internal/backend"
)

var rrInd = 0

func GetNextBackendRR(backends []*backend.Backend) *backend.Backend {
	backend := backends[rrInd%len(backends)]
	rrInd++
	return backend
}
