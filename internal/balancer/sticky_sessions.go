package balancer

import (
	"github.com/lokeshllkumar/load-balancer/internal/backend"
)

func (lb *LoadBalancer) GetBackendSS(clientIP string) *backend.Backend {

	if backendURL, found := lb.clientMap.Load(clientIP); found {
		for _, b := range lb.Pool.GetAliveBackends() {
			if b.URL == backendURL {
				return b
			}
		}
	}
	selected := GetNextBackendRR(lb.Pool.GetAliveBackends())
	lb.clientMap.Store(clientIP, selected.URL)
	return selected
}
